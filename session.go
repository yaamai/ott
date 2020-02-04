package main

import (
    "golang.org/x/crypto/ssh/terminal"
	"github.com/creack/pty"
	"os"
	"os/exec"
//    "go.uber.org/zap"
    "bytes"
//    "time"
    "log"
    "sync"
)

var (
	MARKER = []byte("###OTT-OTT###")
	LF     = []byte("\n")
	CRLF   = []byte("\r\n")
	SPACE  = []byte(" ")
)


type LockedBuffer struct {
    b bytes.Buffer
    m sync.Mutex
}

func (b *LockedBuffer) Bytes() []byte {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Bytes()
}

func (b *LockedBuffer) Reset() {
    b.m.Lock()
    defer b.m.Unlock()
    b.b.Reset()
}

func (b *LockedBuffer) Read(p []byte) (n int, err error) {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Read(p)
}

func (b *LockedBuffer) Write(p []byte) (n int, err error) {
    b.m.Lock()
    defer b.m.Unlock()
    return b.b.Write(p)
}

type Session struct {
	ptmx *os.File
    buffer *LockedBuffer
    Prompt []byte
    cleanupChan chan bool
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

    cleanupChan := make(chan bool)

	c := exec.Command("sh")
	winsize := pty.Winsize{Rows: 10, Cols: 10}
	ptmx, err := pty.StartWithSize(c, &winsize)
	if err != nil {
		return nil, err
	}

    terminal.MakeRaw(int(ptmx.Fd()))

    b := make([]byte, 1024)
    buffer := new(LockedBuffer)
    go func() {
        // TODO: escape infinit for loop
        for {
            l, err := ptmx.Read(b)
            log.Println("for go read from pty", l, err, b[:l], string(b[:l]))
            if err != nil {
                break
            }
            buffer.Write(b[:l])
        }
    }()

    ptmx.Write(LF)
    prompt := readPrompt(buffer)

    // ptmx.SetDeadline(1 * time.Millisecond)


    // skip first prompt
    // l, err := ptmx.Read(b)
    // log.Println("Read from pty", l, err, b[:l])

    /*
    l, err = ptmx.Write([]byte("echo a &&\\echo b\n"))
    log.Println("write to pty", l, err)
    l, err = ptmx.Read(b)
    log.Println("Read from pty", l, err, b[:l], string(b[:l]))
    */

    /*
    go func() {
        l, err := ptmx.Read(b)
        // zap.S().Debug("Read from pty", l, err)
        log.Println("go read from pty", l, err, b[:l])
        // buffer.Write(b[:l])
        // time.Sleep(100 * time.Millisecond)
    }()
    */


    /*
    go func() {
        l, err := ptmx.Write([]byte("aa echo a\r"))
        log.Println("write to pty", l, err)
        // ptmx.Sync()
        time.Sleep(100 * time.Millisecond)
    }()
    */


	r := Session{ptmx: ptmx, buffer: buffer, Prompt: prompt, cleanupChan: cleanupChan}

	return &r, nil
}

func (r *Session) Cleanup() {
	r.ptmx.Close()
}

func (r *Session) SkipPrompt() []byte {
    b := make([]byte, 1024)
    l, err := r.ptmx.Read(b)
    log.Println("skipprompt, Read from pty", l, err, b[:l])

    return b[:l]
}


func getMarkedCommand(cmd string) []byte {
	cmdBytes := []byte(cmd)

	result := make([]byte, 0, len(cmdBytes)+2)
	result = append(result, cmdBytes...)
	result = append(result, []byte("; PS1=")...)
	result = append(result, MARKER...)
	result = append(result, LF...) // generate marker prompt
	result = append(result, []byte("PS1=$ORIG_PS1")...)
	result = append(result, LF...)

	return result
}

func (r *Session) StripEmptyPrompt(buffer, prompt []byte) []byte {
    pos := bytes.LastIndex(buffer, prompt)
    if pos == -1 {
        return buffer
    }

    return buffer[:pos]
}

func ReadOutput(buffer *LockedBuffer) []byte {
    for {
        log.Println(string(buffer.Bytes()))
        pos := bytes.Index(buffer.Bytes(), MARKER)
        if pos == -1 {
            continue
        }
        log.Println("marker pos", pos)

        b := make([]byte, pos)
        buffer.Read(b)

        return b
    }
}

func (r *Session) ExecuteCommand(cmd string) string {
    r.buffer.Reset()
	r.ptmx.Write(getMarkedCommand(cmd))
    return string(ReadOutput(r.buffer))

    // log.Println("aa")
	// go r.ptmx.Write(getMarkedCommand(cmd))
    // log.Println("aa")
	// go r.ptmx.Write(getMarkedCommand(cmd))
    // for {
    // buffer := make([]byte, 128)
    // l, err := r.ptmx.Read(buffer)
    // log.Println("Read from pty", l, err, buffer[:l])
    // }
    // */

    // for {}
    // prompt := r.SkipPrompt()

    // // send command

    // // read output
	// buffer := make([]byte, 65535)
    // l, err := r.ptmx.Read(buffer)
    // log.Println("Read from pty", l, err, buffer[:l])
    // if err != nil {
    //     return ""
    // }

    // r.SkipPrompt()
    // // strip empty prompt&line
    // output := r.StripEmptyPrompt(buffer[:l], prompt)
    // return string(output)
    // */
    // return ""
	// // write_pos := 0
    // // for {
	// // 	write_pos += l
    // // }
    // // return string(buffer[:write_pos])
    // for {
	//     r.ptmx.Write(getMarkedCommand(cmd))
	//     r.ptmx.Write(getMarkedCommand(cmd))

    //     l, err := r.buffer.ReadString(LF[0])
    //     time.Sleep(100 * time.Millisecond)
    //     log.Println("Read from buffer", l, "err=", err)
    //     if err == nil {
    //         break
    //     }
    // }
	// // make read buffer. (bytes.Buffer.ReadFrom() may block while reading EOF)
	// write_pos := 0

	// idx := 0
	// for idx < 100 {
	// 	l, _ := r.ptmx.Read(buffer[write_pos:])
    //     zap.S().Debug("shell-out", buffer[:write_pos])

	// 	isCmdFinished, b := checkMarker(MARKER, buffer[:write_pos])
	// 	if isCmdFinished {
	// 		return string(b)
	// 	}
	// }

	// // return ""
}
