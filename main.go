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
	"encoding/hex"
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
	return append(append([]byte("\necho -n "), marker...), []byte("\n")...)
}


func calcStartMarker(marker, buffer []byte, markerPos []int) (int, int) {
	// get marker line start-pos from marker-start-pos ( sh-5.0$ echo <marker> )
	markerStartPos := markerPos[2]
	markerLineStartPos := bytes.LastIndex(buffer[:markerStartPos], []byte("\r\n"))
	markerLineStartPos += 2
	markerLineEndPos := bytes.Index(buffer[markerLineStartPos:], []byte("\r\n"))
	log.Println("marker-line", markerLineStartPos, markerLineEndPos, hex.Dump(buffer[markerLineStartPos:markerLineStartPos+markerLineEndPos]))

	return markerLineStartPos, markerLineStartPos+markerLineEndPos
}

func calcEndMarker(marker, buffer []byte, markerPos []int) (int, int) {
	endMarkerPos := markerPos[2] + markerPos[3]
	markerLineStartPos := bytes.LastIndex(buffer[:endMarkerPos], []byte("\r\n"))
	markerLineStartPos += 2
	markerLineEndPos := bytes.Index(buffer[endMarkerPos:], []byte("\r\n"))
	return markerLineStartPos, endMarkerPos+markerLineEndPos
}

func getLineByPos(buffer []byte, pos int) (int, int) {
	// get line by pos
	// \r\nhogehoge\r\n
	//       ^ = pos==5
	// -> (2, 11) (hogehoge\r\n)

	lineStart := bytes.LastIndex(buffer[:pos], []byte("\r\n"))
	lineEnd := bytes.Index(buffer[pos:], []byte("\r\n"))
	log.Println("getLineByPos", lineStart, lineEnd)
	return lineStart+2, pos+lineEnd+2
}

func checkMarker(marker, buffer []byte) (bool, []byte) {
	search_start_pos := 0
	// search marker in
        // `echo <marker-echo> \n command \n <marker-echo> <marker> command-output <marker>`
	marker_pos := make([]int, 4)
	for i := 0; i < 4; i++ {
		pos := bytes.Index(buffer[search_start_pos:], marker)
		if pos == -1 {
			log.Println(i)
			return false, nil
		}
		log.Println("check", search_start_pos, pos)
		marker_pos[i] = search_start_pos + pos
		search_start_pos = search_start_pos + pos + len(marker) + 1
	}


	// get target cmdline
	// output_start_pos := 0
	// log.Println(marker_pos)
	// log.Println(hex.Dump(buffer))

	_, outputStart := getLineByPos(buffer, marker_pos[2])
	outputEnd, _ := getLineByPos(buffer, marker_pos[3])
	return true, buffer[outputStart:outputEnd]
}

func (r *Runner) ExecuteCommand(cmd string) {
	MARKER := []byte("### OTT-OTT ###")
	r.ptmx.Write(getMarkerCommand(MARKER))
	r.ptmx.Write([]byte(cmd))
	r.ptmx.Write(getMarkerCommand(MARKER))

	buffer := make([]byte, 65535)
	write_pos := 0
	// marker_pos := 0
	//pre_end_marker_pos := -1
	//post_end_marker_pos := -1
	idx := 0
	for idx < 100 {
		l, _ := r.ptmx.Read(buffer[write_pos:])
		write_pos += l
		isFinished, b := checkMarker(MARKER, buffer[:write_pos])
		log.Println(string(b))
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
	//log.Println(pre_end_marker_pos, post_end_marker_pos)
	//log.Println(string(buffer))
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

	r.ExecuteCommand("env &&\\\nenv")
}
