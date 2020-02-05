package main

import (
	"bytes"
	"errors"
	"github.com/creack/pty"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"os/exec"
	"time"
)

var (
	MARKER = []byte("###OTT-OTT###")
	LF     = []byte("\n")
)

type Session struct {
	ptmx   *os.File
	buffer *LockedBuffer
	prompt []byte
}

func readPrompt(buffer *LockedBuffer) []byte {
	for retry := 0; retry < 10; retry += 1 {

		buf := buffer.Bytes()
		promptLen := len(buf) / 2
		if promptLen > 1 && bytes.Equal(buf[:promptLen], buf[promptLen:]) {
			return buf[:promptLen]
		}

		time.Sleep(1 * time.Millisecond)
	}
	return nil
}

func NewSession() (*Session, error) {
	r := Session{}

	// prepare command
	c := exec.Command("sh")

	// launch and attach to pty
	winsize := pty.Winsize{Rows: 10, Cols: 10}
	ptmx, err := pty.StartWithSize(c, &winsize)
	if err != nil {
		return nil, err
	}
	terminal.MakeRaw(int(ptmx.Fd()))
	r.ptmx = ptmx

	// prepare shell output buffer
	buffer := new(LockedBuffer)
	buffer.Grow(4096)
	r.buffer = buffer

	// wait first prompt
	go r.Reader()
	ptmx.Write(LF)
	prompt := readPrompt(buffer)
	if prompt == nil {
		return nil, errors.New("prompt wait timeout")
	}
	r.prompt = prompt

	return &r, nil
}

func (s *Session) Reader() {
	// TODO: escape infinit for loop
	b := make([]byte, 1024)
	for {
		l, err := s.ptmx.Read(b)
		zap.S().Debug("for go read from pty", l, err, b[:l], string(b[:l]))
		if err != nil {
			break
		}
		s.buffer.Write(b[:l])
	}
}

func (s *Session) Cleanup() {
	s.ptmx.Close()
}

func (s *Session) GetPrompt() []byte {
	return s.prompt
}

func getMarkedCommand(cmd string) []byte {
	cmdBytes := []byte(cmd)

	result := make([]byte, 0, len(cmdBytes)+2)
	result = append(result, cmdBytes...)
	result = append(result, []byte("; PS1=")...)
	result = append(result, MARKER...)
	result = append(result, LF...) // generate marker prompt

	return result
}

func (s *Session) ExecuteCommand(cmd string) string {
	s.buffer.Reset()
	s.ptmx.Write(getMarkedCommand(cmd))
	output, err := s.buffer.ReadToPattern(MARKER)
	if err != nil {
		return ""
	}
	return string(output)
}
