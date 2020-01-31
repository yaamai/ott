package main

import (
	"github.com/creack/pty"
	"log"
	"os"
	"os/exec"
)

var (
	MARKER = []byte("### OTT-OTT ###")
	LF     = []byte("\n")
	SPACE  = []byte(" ")
)

type Session struct {
	ptmx *os.File
}

func NewSession() (*Session, error) {
	c := exec.Command("sh")
	winsize := pty.Winsize{Rows: 10, Cols: 10}
	ptmx, err := pty.StartWithSize(c, &winsize)
	if err != nil {
		return nil, err
	}

	r := Session{ptmx: ptmx}

	return &r, nil
}

func (r *Session) Cleanup() {
	r.ptmx.Close()
}

func getMarkedCommand(cmd string) []byte {
	cmdBytes := []byte(cmd)

	result := make([]byte, 0, len(cmdBytes) + len(MARKER)*2 + 16)
	result = append(result, cmdBytes...)
	result = append(result, SPACE...)
	result = append(result, MARKER...)
	result = append(result, []byte("\necho -n ")...)
	result = append(result, MARKER...)
	result = append(result, LF...)

	return result
}

func (r *Session) ExecuteCommand(cmd string) {
	r.ptmx.Write(getMarkedCommand(cmd))

	// TODO: replace by bytes builder?
	buffer := make([]byte, 65535)
	write_pos := 0

	idx := 0
	for idx < 100 {
		l, _ := r.ptmx.Read(buffer[write_pos:])
		write_pos += l

		isFinished, b := checkMarker(MARKER, buffer[:write_pos])
		log.Println(string(b))
		if isFinished {
			break
		}
	}
}
