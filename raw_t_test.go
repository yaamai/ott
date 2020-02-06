package main

import (
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

func TestRawTParse(t *testing.T) {
	tests := []struct {
		t   string
		l   []Line
		err error
	}{
		{"", []Line{}, nil},
		{"#a", []Line{&CommentLine{"#a"}}, nil},
		{"# meta:", []Line{&MetaCommentLine{"# meta:", nil}}, nil},
		{"# meta:\n#  a: 100", []Line{&MetaCommentLine{"# meta:", nil}, &MetaCommentLine{"#  a: 100", nil}}, nil},
		{"# meta:\n#  a: 100\naaaa:", []Line{&MetaCommentLine{"# meta:", nil}, &MetaCommentLine{"#  a: 100", nil}, &TestCaseLine{"aaaa:"}}, nil},
		{"aaaa:\n  # a", []Line{&TestCaseLine{"aaaa:"}, &TestCaseCommentLine{"  # a", nil}}, nil},
		{"aaaa:\n  $ a", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a", nil}}, nil},
		{"aaaa:\n  $ a\n  a", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a", nil}, &OutputLine{"  a", nil}}, nil},
		{"aaaa:\n  $ a\n  a\n  $ b\n  > c", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a", nil}, &OutputLine{"  a", nil}, &CommandLine{"  $ b", nil}, &CommandContinueLine{"  > c", nil}}, nil},
		{"aaaa:\n  $ a\n  a\n  $ b\n  > c\n  b\n  c", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a", nil}, &OutputLine{"  a", nil}, &CommandLine{"  $ b", nil}, &CommandContinueLine{"  > c", nil}, &OutputLine{"  b", nil}, &OutputLine{"  c", nil}}, nil},
	}
	for _, tt := range tests {
		lines, err := ParseRawT(strings.NewReader(tt.t))

		if err != tt.err {
			t.Fatalf("want = %s, got = %s (%s)", tt.err, err, tt.t)
		}

		if !cmp.Equal(lines, tt.l) {
			t.Fatalf("want = %s, got = %s (%s)", tt.l, lines, tt.t)
		}
	}
}
