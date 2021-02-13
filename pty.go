package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// ShellSession represents shell running session
type ShellSession struct {
	ShellSessionOption
	ptmx   *os.File
	reader *Reader
}

// ShellSessionOption represents shell running options
type ShellSessionOption struct {
	marker     [][]byte
	cmd        *exec.Cmd
	preMarker  []byte
	preCommand []byte
	winsize    pty.Winsize
	buffer     int
	mirror     io.Writer
	parser     ShellParser
}

// Cmd sets command to execute in ShellSession
func Cmd(c *exec.Cmd) func(s *ShellSessionOption) {
	return func(s *ShellSessionOption) {
		s.cmd = c
	}
}

// Mirror sets output mirroring destination for ShellSession
func Mirror(w io.Writer) func(s *ShellSessionOption) {
	return func(s *ShellSessionOption) {
		s.mirror = w
	}
}

// DefaultShellSessionOption returns default ShellSessionOption
func DefaultShellSessionOption() ShellSessionOption {
	marker := [][]byte{[]byte("###OTT"), []byte("OTT###")}
	cmd := exec.Command("sh")
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

// NewShellSession creates ShellSession with opts
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
	sess.parser = NewShellParser(sess.marker, sess.mirror)
	return sess, nil
}

// Reader is flexible byte array reader
type Reader struct {
	base       io.Reader
	buf        []byte
	rpos, wpos int
}

// NewReader creates Reader with underlay io.Reader
func NewReader(size int, r io.Reader) *Reader {
	return &Reader{
		base: r,
		buf:  make([]byte, size),
	}
}

// ReadWithFunc reads bytes array with specified condition
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

// ReadToPattern reads bytes array to specified patterns founds
func (r *Reader) ReadToPattern(pattern []byte) []byte {
	return r.ReadWithFunc(func(buf []byte, l int) (int, []byte) {
		pos := bytes.Index(buf, pattern)
		if pos == -1 {
			return 0, nil
		}

		return pos, buf[:pos]
	})
}

func indexMultiple(buf []byte, patterns ...[][]byte) [][][2]int {
	result := [][][2]int{}
	bufPos := 0

	for _, pattern := range patterns {
		ptnResult := [][2]int{}
		for _, ptn := range pattern {
			pos := bytes.Index(buf[bufPos:], ptn)
			if pos == -1 {
				return result
			}
			ptnResult = append(ptnResult, [2]int{bufPos + pos, bufPos + pos + len(ptn)})
			bufPos += pos + len(ptn)
		}
		if len(ptnResult) == 0 {
			continue
		}
		result = append(result, ptnResult)
	}
	return result
}

// MultiPatternParser is pattern based bytes array parser
type MultiPatternParser struct {
	startPattern, endPattern     [][]byte
	dataCb                       func(data []byte)
	startPatternCb, endPatternCb func(data []byte, pos [][2]int)
}

// Parse implements Reader's Parse function
func (p *MultiPatternParser) Parse(buf []byte, l int) (int, []byte) {
	pos := indexMultiple(buf, p.startPattern, p.endPattern)
	startPos := -1
	if len(pos) > 0 {
		startPos = pos[0][len(pos[0])-1][1]
	}
	endPos := -1
	if len(pos) > 1 {
		endPos = pos[1][0][0]
	}

	p.callDataCallback(buf, l, startPos, endPos)
	p.callPatternCallback(buf, pos)

	if startPos == -1 || endPos == -1 {
		return 0, nil
	}
	return endPos - startPos, buf[startPos:endPos]
}

// TODO: if read partial pattern, callback may incorrect
func (p *MultiPatternParser) callDataCallback(buf []byte, l int, startPos int, endPos int) {
	if p.dataCb == nil {
		return
	}

	cbStartPos := len(buf) - l
	if startPos != -1 && cbStartPos < startPos {
		cbStartPos = startPos
	}
	if startPos != -1 && endPos == -1 {
		if p.dataCb != nil && len(buf) > startPos {
			p.dataCb(buf[cbStartPos:])
		}
	}

	if endPos != -1 && endPos > cbStartPos {
		p.dataCb(buf[cbStartPos:endPos])
	}
}

func (p *MultiPatternParser) callPatternCallback(buf []byte, pos [][][2]int) {
	if len(pos) > 0 && p.startPatternCb != nil {
		p.startPatternCb(buf, pos[0])
	}
	if len(pos) > 1 && p.endPatternCb != nil {
		p.endPatternCb(buf, pos[1])
	}
}

// ShellParser is shell-output adjusted MultiPatternParser
type ShellParser struct {
	MultiPatternParser
	rc     int
	mirror io.Writer
}

// NewShellParser creates ShellParser
func NewShellParser(marker [][]byte, mirror io.Writer) ShellParser {
	p := ShellParser{MultiPatternParser: MultiPatternParser{startPattern: marker, endPattern: marker}, mirror: mirror}
	p.dataCb = p.mirrorData
	p.endPatternCb = p.parseReturnCode
	return p
}

func (p *ShellParser) mirrorData(data []byte) {
	if p.mirror != nil {
		p.mirror.Write(data)
	}
}

func (p *ShellParser) parseReturnCode(data []byte, pos [][2]int) {
	i, err := strconv.Atoi(string(data[pos[0][1]:pos[1][0]]))
	if err != nil {
		p.rc = i
	} else {
		p.rc = -1
	}
}

// Run runs command in ShellSession
func (s *ShellSession) Run(cmd string) (int, string) {
	s.ptmx.Write(s.preCommand)
	s.reader.ReadToPattern(s.preMarker)

	s.ptmx.Write([]byte(cmd))
	outputBytes := s.reader.ReadWithFunc(s.parser.Parse)
	return s.parser.rc, strings.TrimSuffix(string(outputBytes), "\n")
}
