package main

import (
	"testing"
)

func TestGetLineByPos(t *testing.T) {
	tests := []struct {
		b []byte
                pos, spos, epos int
                l string
	}{
		{[]byte(""), 0, 0, 0, ""},
		{[]byte("a"), 0, 0, 1, "a"},
		{[]byte("abcdef"), 0, 0, 6, "abcdef"},
		{[]byte("\r\nabcdef\r\n"), 0, 0, 0, ""},
		{[]byte("\r\nabcdef\r\n"), 2, 2, 8, "abcdef"},
		{[]byte("ABCDEF\r\nabcdef\r\n"), 2, 0, 6, "ABCDEF"},
		{[]byte("ABCDEF\r\nabcdef\r\n"), 6, 0, 6, "ABCDEF"},
		{[]byte("ABCDEF\r\nabcdef\r\n"), 9, 8, 14, "abcdef"},
		{[]byte("\r\nhogehoge\r\n"), 5, 2, 10, "hogehoge"},
	}
	for _, tt := range tests {
		gotSpos, gotEpos := getLineByPos(tt.b, tt.pos)
		if gotSpos != tt.spos || gotEpos != tt.epos {
			t.Fatalf("want = %d, %d, got = %d, %d", tt.spos, tt.epos, gotSpos, gotEpos)
		}
		if string(tt.b[gotSpos:gotEpos]) != tt.l {
			t.Fatalf("want = %s, got = %s", tt.l, string(tt.b[gotSpos:gotEpos]))
		}
	}
}
