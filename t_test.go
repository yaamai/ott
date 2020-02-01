package main

import (
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

/*
	f, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		return TFile{}, err
	}
	defer f.Close()
*/
func TestTFileUnmarshalJSON(t *testing.T) {
	tests := []struct {
		j   string
		t   TFile
		err error
	}{
		{`[]`, TFile{}, nil},
		{`[{"type": "comment", "string": "aa"}]`, TFile{[]Lineable{&Comment{"aa"}}}, nil},
		{`[{"type": "testcase", "name": "aa"}]`, TFile{[]Lineable{&TestCase{Name: "aa"}}}, nil},
		{`[{"type": "testcase", "name": "aa", "steps": [{"commands": ["aa"]}]}]`,
			TFile{[]Lineable{&TestCase{Name: "aa", TestSteps: []*TestStep{&TestStep{Commands: []string{"aa"}}}}}}, nil},
	}
	for _, tt := range tests {
		b := []byte(tt.j)
		tFile := TFile{}
		err := json.Unmarshal(b, &tFile)

		if err != tt.err {
			t.Fatalf("want = %s, got = %s (%s)", tt.err, err, tt.j)
		}

		if !cmp.Equal(tFile.Lines, tt.t.Lines) {
			t.Fatalf("want = %s, got = %s (%s)", tt.t, tFile, tt.j)
		}
	}
}

func TestParseTFile(t *testing.T) {
	tests := []struct {
		s   string
		t   string
		err error
	}{
		{"", `[]`, nil},
		{"# comment", `[{"type": "comment", "string": "# comment"}]`, nil},
		{"# comment\n# comment", `[{"type": "comment", "string": "# comment"}, {"type": "comment", "string": "# comment"}]`, nil},
		{"a:", `[{"type": "testcase", "name": "a:"}]`, nil},
		{"a:\n  $ echo a", `[{"type": "testcase", "name": "a:", "steps": [{"commands": ["  $ echo a"]}]}]`, nil},
		{"a:\n  $ echo a &&\\\n  > echo b", `[{"type": "testcase", "name": "a:", "steps": [{"commands": ["  $ echo a &&\\", "  > echo b"]}]}]`, nil},
		{"a:\n  $ echo a &&\\\n  > echo b\n  a\n  b", `[{"type": "testcase", "name": "a:", "steps": [{"commands": ["  $ echo a &&\\", "  > echo b"], "outputs": ["  a", "  b"]}]}]`, nil},
		{"# meta:\n#  a: 100\na:\n  $ echo a &&\\\n  > echo b\n  a\n  b", `[{"type": "testcase", "name": "a:", "metadata": {"string": "# meta:", "meta": {"a": "100"}}, "steps": [{"commands": ["  $ echo a &&\\", "  > echo b"], "outputs": ["  a", "  b"]}]}]`, nil},
	}
	for _, tt := range tests {
		tFile, err := ParseTFile(strings.NewReader(tt.s))
		if err != tt.err {
			t.Fatalf("want = %s, got = %s (%s)", tt.err, err, tt.s)
		}

		tttFile := TFile{}
		json.Unmarshal([]byte(tt.t), &tttFile)
		if !cmp.Equal(tFile.Lines, tttFile.Lines) {
			t.Fatalf("want = %s, got = %s (%s)", tt.t, tFile, tt.s)
		}
	}
}
