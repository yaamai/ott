package main

import (
	"bytes"
	"time"
)

var (
	START_MARKER     = `###OTT-START###`
	START_MARKER_CMD = []byte(`echo -n "` + START_MARKER + `"; `)
	END_MARKER       = `###OTT-END###`
	END_MARKER_CMD   = []byte(`; echo -n "` + END_MARKER + `"`)
	LF               = []byte("\n")
)

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

func getMarkedCommand(cmd string) []byte {
	cmdBytes := []byte(cmd)

	result := make([]byte, 0)
	result = append(result, START_MARKER_CMD...)
	result = append(result, cmdBytes...)
	result = append(result, END_MARKER_CMD...)
	result = append(result, LF...) // generate marker output

	return result
}

		// output, err := s.buffer.ReadBetweenPattern([]byte(START_MARKER), []byte(END_MARKER))
		// output, err := s.buffer.ReadBetweenPattern([]byte(START_MARKER), []byte(END_MARKER))
