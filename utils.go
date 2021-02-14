package main

import "bytes"

func indexMultiple(buf []byte, patterns ...[][]byte) [][][2]int {
	result := [][][2]int{}
	bufPos := 0

	for _, pattern := range patterns {
		ptnResult := [][2]int{}
		for _, ptn := range pattern {
			pos := bytes.Index(buf[bufPos:], ptn)
			if pos == -1 {
				return result
			}
			ptnResult = append(ptnResult, [2]int{bufPos + pos, bufPos + pos + len(ptn)})
			bufPos += pos + len(ptn)
		}
		if len(ptnResult) == 0 {
			continue
		}
		result = append(result, ptnResult)
	}
	return result
}
