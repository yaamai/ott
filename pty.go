package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"golang.org/x/term"
)

type ShellSession struct {
	ShellSessionOption
	ptmx   *os.File
	reader *Reader
}

type ShellSessionOption struct {
	marker     [][]byte
	cmd        *exec.Cmd
	preMarker  []byte
	preCommand []byte
	winsize    pty.Winsize
	buffer     int
	mirror     io.Writer
}

func Cmd(c *exec.Cmd) func(s *ShellSessionOption) {
	return func(s *ShellSessionOption) {
		s.cmd = c
	}
}

func Mirror(w io.Writer) func(s *ShellSessionOption) {
	return func(s *ShellSessionOption) {
		s.mirror = w
	}
}

func DefaultShellSessionOption() ShellSessionOption {
	marker := [][]byte{[]byte("###OTT"), []byte("OTT###")}
	cmd := exec.Command("bash")
	cmd.Env = append(cmd.Env, "PS1="+string(marker[0])+"$?"+string(marker[1]))
	cmd.Env = append(cmd.Env, "PS2=")
	cmd.Env = append(cmd.Env, "HISTFILE=/dev/null")
	cmd.Args = append(cmd.Args, []string{"--norc", "--noprofile"}...)
	preMarker := "###OTT-PRE-OTT###"
	preCommand := "eval \"echo -n '" + preMarker + "'; (exit $?)\"\n"
	winsize := pty.Winsize{Cols: 80, Rows: 24}
	buffer := 65536

	return ShellSessionOption{
		marker:     marker,
		cmd:        cmd,
		preMarker:  []byte(preMarker),
		preCommand: []byte(preCommand),
		winsize:    winsize,
		buffer:     buffer,
	}
}

func NewShellSession(opts ...func(s *ShellSessionOption)) (*ShellSession, error) {
	sess := &ShellSession{}
	sess.ShellSessionOption = DefaultShellSessionOption()
	for _, opt := range opts {
		opt(&sess.ShellSessionOption)
	}

	ptmx, err := pty.StartWithSize(sess.cmd, &sess.winsize)
	if err != nil {
		return nil, err
	}
	sess.ptmx = ptmx
	term.MakeRaw(int(ptmx.Fd()))

	sess.reader = NewReader(sess.buffer, ptmx)
	return sess, nil
}

type Reader struct {
	base       io.Reader
	buf        []byte
	rpos, wpos int
}

func NewReader(size int, r io.Reader) *Reader {
	return &Reader{
		base: r,
		buf:  make([]byte, size),
	}
}

func (r *Reader) ReadWithFunc(f func([]byte, int) (int, []byte)) []byte {
	for {
		l, _ := r.base.Read(r.buf[r.wpos:])
		r.wpos += l

		l, output := f(r.buf[r.rpos:r.wpos], l)
		if output != nil {
			r.rpos = 0
			r.wpos = 0
			return output
		}
	}
}

func (r *Reader) ReadToPattern(pattern []byte) []byte {
	return r.ReadWithFunc(func(buf []byte, l int) (int, []byte) {
		pos := bytes.Index(buf, pattern)
		if pos == -1 {
			return 0, nil
		}

		return pos, buf[:pos]
	})
}

// TODO: support multiple patter
func indexMultiple(buf []byte, patterns... [][]byte) [][2]int {
	result := [][2]int{}
	bufPos := 0
	
	for _, pattern := range patterns {
		for _, ptn := range pattern {
			pos := bytes.Index(buf[bufPos:], ptn)
			if pos == -1 {
				return nil
			}
			result = append(result, [2]int{bufPos + pos, bufPos + pos + len(ptn)})
			bufPos += pos + len(ptn)
		}
	}
	return result
}

func indexPatterns(buf []byte, startPattern, endPattern [][]byte) (int, int, [][2]int, [][2]int) {
	spList := indexMultiple(buf, startPattern)
	if spList == nil {
		return -1, -1, nil, nil
	}
	startPos := spList[len(spList)-1][1]

	epList := indexMultiple(buf[startPos:], endPattern)
	if epList != nil {
		return startPos, startPos + epList[0][0], spList, epList
	}

	return startPos, -1, spList, nil
}

func callCallback(buf []byte, l int, startPos int, endPos int, cb func(data []byte)) {
	if cb == nil {
		return
	}

	cbStartPos := len(buf) - l
	if startPos != -1 && cbStartPos < startPos {
		cbStartPos = startPos
	}
	if startPos != -1 && endPos == -1 {
		if cb != nil && len(buf) > startPos {
			cb(buf[cbStartPos:])
		}
	}

	if endPos != -1 && endPos > cbStartPos {
		cb(buf[cbStartPos:endPos])
	}
}

// TODO: if read partial pattern, callback may incorrect
func readBetweenMultiplePatternFunc(startPattern, endPattern [][]byte, dataCb func(data []byte), endPatternCb func(data []byte, pos [][2]int)) func(buf []byte, l int) (int, []byte) {
	return func(buf []byte, l int) (int, []byte) {
		startPos, endPos, _, endPosList := indexPatterns(buf, startPattern, endPattern)
		callCallback(buf, l, startPos, endPos, dataCb)
		if startPos == -1 || endPos == -1 {
			return 0, nil
		}

		if endPatternCb != nil {
			endPatternCb(buf, endPosList)
		}
		return endPos - startPos, buf[startPos:endPos]
	}
}

func (s *ShellSession) Run(cmd string) (int, string) {
	s.ptmx.Write(s.preCommand)
	s.reader.ReadToPattern(s.preMarker)

	s.ptmx.Write([]byte(cmd))
	// TODO: separate to struct and ReadWith(Interface)
	outputBytes := s.reader.ReadWithFunc(readBetweenMultiplePatternFunc(s.marker, s.marker, func(data []byte) {
		if s.mirror != nil {
			s.mirror.Write(data)
		}
	}, func(data []byte, pos [][2]int) {
		println("end", string(data), pos[0][0], pos[0][1])
	}))
	return 0, strings.TrimSuffix(string(outputBytes), "\n")
}
