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
	START_MARKER     = `###OTT-START###`
	START_MARKER_CMD = []byte(`echo -n "` + START_MARKER + `"; `)
	END_MARKER       = `###OTT-END###`
	END_MARKER_CMD   = []byte(`; echo -n "` + END_MARKER + `"`)
	LF               = []byte("\n")
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
		zap.S().Debug("read from pty", l, err, b[:l], string(b[:l]))
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

	result := make([]byte, 0)
	result = append(result, START_MARKER_CMD...)
	result = append(result, cmdBytes...)
	result = append(result, END_MARKER_CMD...)
	result = append(result, LF...) // generate marker output

	return result
}

func (s *Session) ExecuteCommand(cmd string) string {
	s.buffer.Reset()
	s.ptmx.Write(getMarkedCommand(cmd))
	for retry := 0; retry < 100; retry += 1 {
		output, err := s.buffer.ReadBetweenPattern([]byte(START_MARKER), []byte(END_MARKER))
		zap.S().Debug("wait output", output, err)
		if err != nil {
			return ""
		}
		if output == nil && err == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		return string(output)
	}
	return ""
}
