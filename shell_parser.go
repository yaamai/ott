package main

import (
	"bytes"
)

func getLineByPos(buffer []byte, pos int) (int, int) {
	lineStart := bytes.LastIndex(buffer[:pos], []byte("\r\n"))
	if lineStart != -1 {
		// forward "\r\n" == 2bytes
		lineStart += 2
	}
	lineEnd := bytes.Index(buffer[pos:], []byte("\r\n"))
	if lineEnd != -1 {
		// adjust search base
		lineEnd = pos + lineEnd
	}
	if lineStart == -1 || lineEnd == -1 {
		if lineStart == -1 {
			lineStart = 0
		}
		if lineEnd == -1 {
			lineEnd = len(buffer)
		}
		return lineStart, lineEnd
	}
	return lineStart, lineEnd
}

func checkMarker(marker, buffer []byte) (bool, []byte) {
	search_start_pos := 0
	// search marker in buffer 4-times
	// `echo <marker-echo> \n command <marker>\n <marker-echo> <marker> command-output <marker>`
	marker_pos := make([]int, 4)
	for i := 0; i < 4; i++ {
		pos := bytes.Index(buffer[search_start_pos:], marker)
		if pos == -1 {
			// log.Println(i)
			return false, nil
		}
		// log.Println("check", search_start_pos, pos)
		marker_pos[i] = search_start_pos + pos
		search_start_pos = search_start_pos + pos + len(marker) + 1
	}

	_, outputStart := getLineByPos(buffer, marker_pos[2])
	outputEnd, _ := getLineByPos(buffer, marker_pos[3])
	return true, buffer[outputStart+2 : outputEnd-2]
}
