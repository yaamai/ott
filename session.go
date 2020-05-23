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

const (
	INTERNAL_BUFFER_SIZE = 65536
	READ_BUFFER_SIZE = 4096
	PROMPT_RETRY = 100
	PROMPT_RETRY_WAIT = 10
	CMD_EXECUTE_RETRY = 100
	CMD_EXECUTE_RETRY_WAIT = 10
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
	ptmx    *os.File
	buffer  *LockedBuffer
	adapter SessionAdapter
}

func getSessionAdapter(mode string) SessionAdapter {
	switch mode {
	case "shell":
		return &ShellSession{}
	case "python":
		return &PythonSession{}
	}
	return nil
}

func NewSession(cmd, mode string) (*Session, error) {
	r := Session{}

	// launch and attach to pty
	c := exec.Command(cmd)
	winsize := pty.Winsize{Cols: 80, Rows: 24}
	ptmx, err := pty.StartWithSize(c, &winsize)
	if err != nil {
		return nil, err
	}
	terminal.MakeRaw(int(ptmx.Fd()))
	r.ptmx = ptmx

	r.adapter = getSessionAdapter(mode)
	if r.adapter == nil {
		return nil, errors.New("Unsupported mode specified")
	}

	// prepare shell output buffer
	buffer := new(LockedBuffer)
	buffer.Grow(INTERNAL_BUFFER_SIZE)
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
	b := make([]byte, READ_BUFFER_SIZE)
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

	for retry := 0; retry < CMD_EXECUTE_RETRY; retry += 1 {
		output, err := s.buffer.ReadBetweenPattern(startMarker, endMarker)
		zap.S().Debug("wait output", output, string(output), err, s.buffer.Bytes())
		if err != nil {
			return []string{}
		}
		if output == nil && err == nil {
			time.Sleep(CMD_EXECUTE_RETRY_WAIT * time.Millisecond)
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
	buffer.WaitStable(PROMPT_RETRY, PROMPT_RETRY_WAIT*time.Millisecond)
	time.Sleep(PROMPT_RETRY_WAIT*time.Millisecond)

	// send LF
	buffer.Reset()
	ptmx.Write(LF)

	// recv prompt
	buffer.WaitStable(PROMPT_RETRY, PROMPT_RETRY_WAIT*time.Millisecond)
	firstPrompt := make([]byte, buffer.Len())
	copy(firstPrompt, buffer.Bytes())

	// re-send LF
	buffer.Reset()
	ptmx.Write(LF)

	// recv prompt and compare
	buffer.WaitStable(PROMPT_RETRY, PROMPT_RETRY_WAIT*time.Millisecond)
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
