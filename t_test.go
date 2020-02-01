package main

import (
	"testing"
	"strings"
	"github.com/google/go-cmp/cmp"
)

/*
	f, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		return TFile{}, err
	}
	defer f.Close()
*/
func TestParseTFile(t *testing.T) {
	tests := []struct {
		s    string
		t TFile
		err error
	}{
		{"", TFile{}, nil},
		{"# comment", TFile{[]Lineable{Comment("# comment")}}, nil},
	}
	for _, tt := range tests {
		tFile, err := ParseTFile(strings.NewReader(tt.s))
		if err != tt.err {
			t.Fatalf("want = %s, got = %s (%s)", tt.err, err, tt.s)
		}

		if cmp.Equal(tFile, tt.t) {
			t.Fatalf("want = %s, got = %s (%s)", tt.t, tFile, tt.s)
		}
	}
}

