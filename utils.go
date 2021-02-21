package main

import (
	"bytes"
	"strings"
)

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

func hasAnyPrefix(s string, prefixes []string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return prefix
		}
	}
	return ""
}

func splitByPrefixes(lines []string, sepPrefixes []string, prefixes []string) []map[string][]string {
	result := []map[string][]string{}

	t := map[string][]string{}
	for _, l := range lines {
		p := hasAnyPrefix(l, sepPrefixes)
		if p != "" && len(t) > 0 {
			result = append(result, t)
			t = map[string][]string{}
		}
		p = hasAnyPrefix(l, prefixes)
		t[p] = append(t[p], strings.TrimPrefix(l, p))
	}
	if len(t) > 0 {
		result = append(result, t)
	}

	return result
}
