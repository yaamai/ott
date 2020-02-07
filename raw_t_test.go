package main

import (
	"github.com/stretchr/testify/assert"
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
		{"# meta:", []Line{&MetaCommentLine{"# meta:"}}, nil},
		{"# meta:\n#  a: 100", []Line{&MetaCommentLine{"# meta:"}, &MetaCommentLine{"#  a: 100"}}, nil},
		{"# meta:\n#  a: 100\naaaa:", []Line{&MetaCommentLine{"# meta:"}, &MetaCommentLine{"#  a: 100"}, &TestCaseLine{"aaaa:"}}, nil},
		{"aaaa:\n  # a", []Line{&TestCaseLine{"aaaa:"}, &TestCaseCommentLine{"  # a"}}, nil},
		{"aaaa:\n  $ a", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a"}}, nil},
		{"aaaa:\n  $ a\n  a", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a"}, &OutputLine{"  a"}}, nil},
		{"aaaa:\n  $ a\n  a\n  $ b\n  > c", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a"}, &OutputLine{"  a"}, &CommandLine{"  $ b"}, &CommandContinueLine{"  > c"}}, nil},
		{"aaaa:\n  $ a\n  a\n  $ b\n  > c\n  b\n  c", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a"}, &OutputLine{"  a"}, &CommandLine{"  $ b"}, &CommandContinueLine{"  > c"}, &OutputLine{"  b"}, &OutputLine{"  c"}}, nil},
	}
	for _, tt := range tests {
		lines, err := ParseRawT(strings.NewReader(tt.t))

		if err != tt.err {
			t.Fatalf("want = %s, got = %s (%s)", tt.err, err, tt.t)
		}

		assert.Equal(t, tt.l, lines)
	}
}
