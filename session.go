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

		// remove LF and strip DSR for alpine+busybox+sh
		cleanedBuf := bytes.ReplaceAll(bytes.ReplaceAll(buf, LF, []byte{}), []byte{27, 91, 54, 110}, []byte{})
		promptLen = len(cleanedBuf) / 2
		if promptLen > 1 && bytes.Equal(cleanedBuf[:promptLen], cleanedBuf[promptLen:]) {
			return buf[:promptLen]
		}

		time.Sleep(10 * time.Millisecond)
	}
	return nil
}


func waitBufferStable(buffer *LockedBuffer) {
	oldlen := buffer.Len()
	count := 0
	waitTime := 0
	for retry := 0; retry < 100; retry += 1 {
		l := buffer.Len()
		if oldlen > 0 && count > 5 {
			break
		}
		if l == oldlen {
			count += 1
		}
		oldlen = l
		time.Sleep(10 * time.Millisecond)
		waitTime += 10
	}

	zap.S().Debug("Waited!", waitTime)
}

func guessPrompt(buffer *LockedBuffer, ptmx *os.File) []byte {
	// wait first non-empty read, and wait continued data
	waitBufferStable(buffer)
	time.Sleep(10 * time.Millisecond)

	// send LF
	buffer.Reset()
	ptmx.Write(LF)

	// recv prompt
	waitBufferStable(buffer)
	firstPrompt := make([]byte, buffer.Len())
	copy(firstPrompt, buffer.Bytes())

	// re-send LF
	buffer.Reset()
	ptmx.Write(LF)

	// recv prompt and compare
	waitBufferStable(buffer)
	secondPrompt := make([]byte, buffer.Len())
	copy(secondPrompt, buffer.Bytes())

	zap.S().Debug("prompt: ", bytes.Equal(firstPrompt, secondPrompt), firstPrompt, secondPrompt)
	if bytes.Equal(firstPrompt, secondPrompt) {
		return firstPrompt
	}
	return nil
}

func NewSession() (*Session, error) {
	r := Session{}

	// prepare command
	c := exec.Command("python")

	// launch and attach to pty
	winsize := pty.Winsize{Rows: 50, Cols: 50}
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
	prompt := guessPrompt(buffer, ptmx)
	if prompt == nil {
		return nil, errors.New("prompt wait timeout")
	}
	r.prompt = bytes.TrimSuffix(bytes.TrimPrefix(prompt, []byte("\r\n")), []byte("\r\n"))

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

func (s *Session) ExecuteCommand(cmdStrs []string) []string {
	s.buffer.Reset()

	// generate cmdline and wait-pattern
	cmds := getBytesArray(cmdStrs)
	cmdline := bytes.Join(cmds, []byte("\n"))
	cmdline = append(cmdline, []byte("\n\n")...)
	expectStartMarker := bytes.Join(cmds, []byte("\r\n... "))
	expectStartMarker = append(expectStartMarker, []byte("\r\n")...)

	s.ptmx.Write(cmdline)
	zap.S().Debug("Execute: ", string(cmdline), cmdline, expectStartMarker)
	for retry := 0; retry < 100; retry += 1 {
		// output, err := s.buffer.ReadBetweenPattern([]byte(START_MARKER), []byte(END_MARKER))
		// output, err := s.buffer.ReadBetweenPattern([]byte(START_MARKER), []byte(END_MARKER))
		zap.S().Debug("Buf", s.buffer.Bytes())
		output, err := s.buffer.ReadBetweenPattern(expectStartMarker, s.prompt)
		zap.S().Debug("wait output", output, err)
		if err != nil {
			return []string{}
		}
		if output == nil && err == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		return getStringArray(bytes.Split(output, []byte("\r\n")))
	}
	return []string{}
}
