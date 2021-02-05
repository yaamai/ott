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
	marker     []byte
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
	marker := "###OTT###"
	markerCmd := "###OTT-$?###"
	cmd := exec.Command("bash")
	cmd.Env = append(cmd.Env, "PS1="+markerCmd)
	cmd.Env = append(cmd.Env, "PS2=")
	cmd.Env = append(cmd.Env, "HISTFILE=/dev/null")
	cmd.Args = append(cmd.Args, []string{"--norc", "--noprofile"}...)
	preMarker := "###OTT-PRE###"
	preCommand := "eval \"echo -n '" + preMarker + "'; (exit $?)\"\n"
	winsize := pty.Winsize{Cols: 80, Rows: 24}
	buffer := 65536

	return ShellSessionOption{
		marker:     []byte(marker),
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
			r.rpos += l
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

func getPatternPos(buf, startPattern, endPattern []byte) (int, int) {
	startPos := bytes.Index(buf, startPattern)
	if startPos == -1 {
		return -1, -1
	}
	startPos += len(startPattern)

	endPos := bytes.Index(buf[startPos:], endPattern)
	if endPos == -1 {
		return startPos, -1
	}
	endPos += startPos

	return startPos, endPos
}

func callCallback(cb func(data []byte)) {
	if cb == nil {
		return
	}

	cbStartPos := len(buf)-l
	if startPos != -1 && cbStartPos < startPos {
		cbStartPos = startPos
	}
	if startPos != -1 && endPos == -1 {
		if cb != nil && len(buf) > startPos {
			cb(buf[cbStartPos:])
		}
	}

	if cb != nil && l > len(endPattern) {
		cb(buf[cbStartPos: len(buf)-len(endPattern)])
	}
}

// TODO: if read partial pattern, callback may incorrect
func (r *Reader) ReadBetweenPattern(startPattern, endPattern []byte, cb func(data []byte)) []byte {
	return r.ReadWithFunc(func(buf []byte, l int) (int, []byte) {
		startPos, endPos := getPatternPos(buf, startPattern, endPattern)
		callCallback(cb)
		if startPos == -1 || endPos == -1 {
			return 0, nil
		}

		return endPos - startPos, buf[startPos:endPos]
	})
}

func (s *ShellSession) Run(cmd string) string {
	s.ptmx.Write(s.preCommand)
	s.reader.ReadToPattern(s.preMarker)

	s.ptmx.Write([]byte(cmd))
	outputBytes := s.reader.ReadBetweenPattern(s.marker, s.marker, func(data []byte) {
		if s.mirror != nil {
			s.mirror.Write(data)
		}
	})
	return strings.TrimSuffix(string(outputBytes), "\n")
}
