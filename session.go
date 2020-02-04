package main

import (
    "golang.org/x/crypto/ssh/terminal"
	"github.com/creack/pty"
	"os"
	"os/exec"
    "go.uber.org/zap"
    "bytes"
)

var (
	MARKER = []byte("###OTT-OTT###")
	LF     = []byte("\n")
)


type Session struct {
	ptmx *os.File
    buffer *LockedBuffer
    prompt []byte
}


func readPrompt(buffer *LockedBuffer) []byte {
    // TODO: add timeout and break
    for {
        buf := buffer.Bytes()
        promptLen := len(buf) / 2
        // log.Println("promptlens", promptLen, buf[:promptLen])
        if promptLen < 1 {
            continue
        }
        if bytes.Equal(buf[:promptLen], buf[promptLen:]) {
            return buf[:promptLen]
        }
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
    output := s.buffer.ReadToPattern(MARKER)
    return string(output)
}
