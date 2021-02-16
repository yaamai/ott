package main

import "bytes"

func indexMultiple(buf []byte, patterns ...[][]byte) [][][2]int64 {
	result := [][][2]int64{}
	bufPos := 0

	for _, pattern := range patterns {
		ptnResult := [][2]int64{}
		for _, ptn := range pattern {
			pos := bytes.Index(buf[bufPos:], ptn)
			if pos == -1 {
				return result
			}
			ptnResult = append(ptnResult, [2]int64{int64(bufPos + pos), int64(bufPos + pos + len(ptn))})
			bufPos += pos + len(ptn)
		}
		if len(ptnResult) == 0 {
			continue
		}
		result = append(result, ptnResult)
	}
	return result
}
