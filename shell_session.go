package main

import (
	"io"
	"os"
	"os/exec"
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
	parser     *ShellParser
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
	buffer := 128

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

// Run runs command in ShellSession
func (s *ShellSession) Run(cmd string) (int, string) {
	s.ptmx.Write(s.preCommand)
	s.reader.ReadToPattern(s.preMarker)

	s.ptmx.Write([]byte(cmd))
	outputBytes := s.reader.ReadWithFunc(s.parser.Parse)
	return s.parser.rc, strings.TrimSuffix(string(outputBytes), "\n")
}
