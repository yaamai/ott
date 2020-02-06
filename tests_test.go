package main

import (
	"testing"
	"github.com/google/go-cmp/cmp"
)

func TestNewFromRawT(t *testing.T) {
	tests := []struct {
		l   []Line
		t   TestFile
	}{
		{[]Line{}, TestFile{}},
		{[]Line{&CommentLine{"#a"}}, TestFile{Comments: []string{"#a"}}},
		{[]Line{&MetaCommentLine{"# meta:", nil}}, TestFile{}},
		{[]Line{&MetaCommentLine{"# meta:", nil}, &MetaCommentLine{"#  a: 100", nil}}, TestFile{Metadata: map[string]string{a: 100}}},
		// {"# meta:\n#  a: 100\naaaa:", []Line{&MetaCommentLine{"# meta:", nil}, &MetaCommentLine{"#  a: 100", nil}, &TestCaseLine{"aaaa:"}}, nil},
		// {"aaaa:\n  # a", []Line{&TestCaseLine{"aaaa:"}, &TestCaseCommentLine{"  # a", nil}}, nil},
		// {"aaaa:\n  $ a", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a", nil}}, nil},
		// {"aaaa:\n  $ a\n  a", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a", nil}, &OutputLine{"  a", nil}}, nil},
		// {"aaaa:\n  $ a\n  a\n  $ b\n  > c", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a", nil}, &OutputLine{"  a", nil}, &CommandLine{"  $ b", nil}, &CommandContinueLine{"  > c", nil}}, nil},
		// {"aaaa:\n  $ a\n  a\n  $ b\n  > c\n  b\n  c", []Line{&TestCaseLine{"aaaa:"}, &CommandLine{"  $ a", nil}, &OutputLine{"  a", nil}, &CommandLine{"  $ b", nil}, &CommandContinueLine{"  > c", nil}, &OutputLine{"  b", nil}, &OutputLine{"  c", nil}}, nil},
	}
	for _, tt := range tests {
        testFile := NewFromRawT(tt.l)

		if !cmp.Equal(testFile, tt.t) {
			t.Fatalf("want = %s, got = %s (%s)", tt.t, testFile, tt.l)
		}
	}
}
