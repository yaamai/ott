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

type SessionAdapter interface {
	GuessPrompt(*LockedBuffer, *os.File) []byte
	GetPrompt() []byte
	GetCmdline([]string) []byte
	GetStartMarker([]string) []byte
	GetEndMarker([]string) []byte
	NormalizeOutput([]byte) []string
}

type Session struct {
	ptmx   *os.File
	buffer *LockedBuffer
	adapter SessionAdapter
}

func NewSession() (*Session, error) {
	r := Session{}


	// launch and attach to pty
	c := exec.Command("sh")
	winsize := pty.Winsize{Rows: 50, Cols: 50}
	ptmx, err := pty.StartWithSize(c, &winsize)
	if err != nil {
		return nil, err
	}
	terminal.MakeRaw(int(ptmx.Fd()))
	r.ptmx = ptmx
	r.adapter = &ShellSession{}

	// prepare shell output buffer
	buffer := new(LockedBuffer)
	buffer.Grow(4096)
	r.buffer = buffer

	// wait first prompt
	go r.Reader()
	prompt := r.adapter.GuessPrompt(buffer, ptmx)
	if prompt == nil {
		return nil, errors.New("prompt wait timeout")
	}

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

func (s *Session) ExecuteCommand(cmdStrs []string) []string {
	s.buffer.Reset()

	// generate cmdline and wait-pattern
	cmdline := s.adapter.GetCmdline(cmdStrs)
	startMarker := s.adapter.GetStartMarker(cmdStrs)
	endMarker := s.adapter.GetEndMarker(cmdStrs)
	s.ptmx.Write(cmdline)
	zap.S().Debug("Execute: ", string(cmdline), cmdline)

	for retry := 0; retry < 100; retry += 1 {
		output, err := s.buffer.ReadBetweenPattern(startMarker, endMarker)
		zap.S().Debug("wait output", output, err, s.buffer.Bytes())
		if err != nil {
			return []string{}
		}
		if output == nil && err == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		return s.adapter.NormalizeOutput(output)
	}
	return []string{}
}

func (s *Session) Cleanup() {
	s.ptmx.Close()
}

func (s *Session) GetPrompt() []byte {
	return s.adapter.GetPrompt()
}

func guessPrompt(buffer *LockedBuffer, ptmx *os.File) []byte {
	// wait first non-empty read, and wait continued data
	buffer.WaitStable(100, time.Millisecond*10)
	time.Sleep(10 * time.Millisecond)

	// send LF
	buffer.Reset()
	ptmx.Write(LF)

	// recv prompt
	buffer.WaitStable(100, time.Millisecond*10)
	firstPrompt := make([]byte, buffer.Len())
	copy(firstPrompt, buffer.Bytes())

	// re-send LF
	buffer.Reset()
	ptmx.Write(LF)

	// recv prompt and compare
	buffer.WaitStable(100, time.Millisecond*10)
	secondPrompt := make([]byte, buffer.Len())
	copy(secondPrompt, buffer.Bytes())

	zap.S().Debug("prompt: ", bytes.Equal(firstPrompt, secondPrompt), firstPrompt, secondPrompt)
	if bytes.Equal(firstPrompt, secondPrompt) {
		return firstPrompt
	}
	return nil
}

func getBytesArray(array []string) [][]byte {
	bytesArray := [][]byte{}
	for idx, _ := range array {
		bytesArray = append(bytesArray, []byte(array[idx]))
	}

	return bytesArray
}

func getStringArray(array [][]byte) []string {
	stringArray := []string{}
	for idx, _ := range array {
		stringArray = append(stringArray, string(array[idx]))
	}

	return stringArray
}
