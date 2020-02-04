package main

import (
	"github.com/creack/pty"
//    "io"
	"os"
	"os/exec"
    "go.uber.org/zap"
    "bytes"
)

var (
	MARKER = []byte("### OTT-OTT ###")
	LF     = []byte("\n")
	SPACE  = []byte(" ")
)

type Session struct {
	ptmx *os.File
    buffer *bytes.Buffer
}

func NewSession() (*Session, error) {
	c := exec.Command("sh")
	winsize := pty.Winsize{Rows: 10, Cols: 10}
	ptmx, err := pty.StartWithSize(c, &winsize)
	if err != nil {
		return nil, err
	}

    buffer := new(bytes.Buffer)
    buffer.Grow(65535)
    go func() {
        zap.S().Debug("start buffering")
        for {
            b := make([]byte, 1024)
            l, _ := ptmx.Read(b)
            buffer.Write(b[:l])
            zap.S().Debug(b[:l])
            // io.Copy(buffer, ptmx)
        //     l, err := buffer.ReadFrom(ptmx)
            zap.S().Debug("bytes.Buffer")
        }
    }()

    // dirty fix to echo-back on first terminal read
    // ptmx.Write([]byte(":\n"))
	// buffer := make([]byte, 65535)
    // ptmx.Read(buffer)
    // ptmx.Read(buffer)
    // ptmx.Read(buffer)

	r := Session{ptmx: ptmx, buffer: buffer}

	return &r, nil
}

func (r *Session) Cleanup() {
	r.ptmx.Close()
}


func getMarkedCommand(cmd string) []byte {
	cmdBytes := []byte(cmd)

	result := make([]byte, 0, len(cmdBytes)+len(MARKER)*2+16)
	result = append(result, LF...)
	result = append(result, cmdBytes...)
	result = append(result, SPACE...)
	result = append(result, MARKER...)
	result = append(result, []byte("\necho -n ")...)
	result = append(result, MARKER...)
	result = append(result, LF...)

	return result
}

func (r *Session) ExecuteCommand(cmd string) string {
	r.ptmx.Write(getMarkedCommand(cmd))

    for {
        s, err := r.buffer.ReadString(LF[0])
        zap.S().Debug("ReadString, ", s, err)
        if err != nil {
            break
        }
    }

	// make read buffer. (bytes.Buffer.ReadFrom() may block while reading)
    /*
	write_pos := 0

	idx := 0
	for idx < 100 {
		l, _ := r.ptmx.Read(buffer[write_pos:])
		write_pos += l
        zap.S().Debug("shell-out", buffer[:write_pos])

		isCmdFinished, b := checkMarker(MARKER, buffer[:write_pos])
		if isCmdFinished {
			return string(b)
		}
	}
    */
	return ""
}
