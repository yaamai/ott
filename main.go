package main

import (
	//        "io"
	"log"
	"os"
	"os/exec"
	//        "os/signal"
	//        "syscall"

	"github.com/creack/pty"
	//        "golang.org/x/crypto/ssh/terminal"

//	"strings"
	"bytes"
)


// TODO:
// Session
// ShellParser
// Runner
// TestFile
// TestCase
// TestStep

type Runner struct {
	ptmx *os.File
}

func NewRunner() (*Runner, error) {
	c := exec.Command("sh")
	winsize := pty.Winsize{Rows: 10, Cols: 10}
	ptmx, err := pty.StartWithSize(c, &winsize)
	if err != nil {
		return nil, err
	}

	r := Runner{ptmx: ptmx}

	return &r, nil
}


func getMarkerCommand(marker []byte) []byte {
	return append(append([]byte("\n"), marker...), []byte("\n")...)
}

func checkMarker(startMarker, endMarker, buffer []byte) (bool, []byte) {
	search_start_pos := 0
	// to skip echo-back marker
	pos := bytes.Index(buffer[search_start_pos:], []byte(startMarker))
	if pos == -1 {
		return false, nil
	}
	search_start_pos = pos + 1
	pos = bytes.Index(buffer[search_start_pos:], []byte(startMarker))
	if pos == -1 {
		return false, nil
	}
	search_start_pos = pos + 1
	pos = bytes.Index(buffer[search_start_pos:], []byte(endMarker))
	if pos == -1 {
		return false, nil
	}

	return true, buffer
}

func (r *Runner) ExecuteCommand(cmd string) {
	START_MARKER := []byte("### OTT-START ###")
	END_MARKER := []byte("### OTT-END ###")
	r.ptmx.Write(getMarkerCommand(START_MARKER))
	r.ptmx.Write([]byte(cmd))
	r.ptmx.Write(getMarkerCommand(END_MARKER))

	buffer := make([]byte, 65535)
	write_pos := 0
	// marker_pos := 0
	pre_end_marker_pos := -1
	post_end_marker_pos := -1
	idx := 0
	for idx < 100 {
		l, _ := r.ptmx.Read(buffer[write_pos:])
		write_pos += l
		isFinished, _ := checkMarker(START_MARKER, END_MARKER, buffer[:write_pos])
		if isFinished {
			break
		}
		//marker_pos := strings.Index(string(buffer)[start_marker_pos+15:], "====")
		//if marker_pos != -1 {
		//	if start_marker_pos != 0{
		//		break
		//	}
		//	start_marker_pos = marker_pos
		//}
		//log.Println(string(buffer))
		//log.Println((buffer[0:write_pos]))
		// if pre_end_marker_pos != -1 && post_end_marker_pos != -1 {
		// 	break
		// }
		// //log.Println("targ", buffer[marker_pos:write_pos])
		// pos := bytes.Index(buffer[marker_pos:], []byte(END_MARKER))
		// //log.Println("pos", pos, len(END_MARKER))
		// if pre_end_marker_pos == -1 && pos != -1 {
		// 	pre_end_marker_pos = pos
		// 	marker_pos = pos + 1
		// 	log.Println("Found pre", marker_pos)
		// 	continue
		// }
		// if pre_end_marker_pos != -1 && pos != -1 {
		// 	log.Println("Foudn post", marker_pos)
		// 	post_end_marker_pos = pos
		// 	//break
		// }
		idx += 1
	}
	log.Println(pre_end_marker_pos, post_end_marker_pos)
	log.Println(string(buffer))
}

func (r *Runner) read() {
	func() {
		for {
			b := make([]byte, 1024)
			l, _ := r.ptmx.Read(b)
			log.Println("===> ", l)
			log.Println("===> ", string(b))
		}
	}()
}

func (r *Runner) Cleanup() {
	r.ptmx.Close()
}

func main() {
	r, err := NewRunner()
	if err != nil {
		log.Fatalln(err)
	}

	r.ExecuteCommand("env")
}
