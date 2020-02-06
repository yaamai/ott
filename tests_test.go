package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFromRawT(t *testing.T) {
	tests := []struct {
		l []Line
		t TestFile
	}{
		{
			[]Line{},
			TestFile{},
		},
		{
			[]Line{&CommentLine{"#a"}},
			TestFile{Comments: []string{"#a"}},
		},
		{
			[]Line{&MetaCommentLine{"# meta:", nil}},
			TestFile{},
		},
		{
			[]Line{
				&MetaCommentLine{"# meta:", nil},
				&MetaCommentLine{"#  a: 100", nil},
			},
			TestFile{},
		},
		{
			[]Line{
				&MetaCommentLine{"# meta:", nil},
				&MetaCommentLine{"#  a: 100", nil},
				&TestCaseLine{"aaaa:"},
			},
			TestFile{
				Tests: []TestCase{
					TestCase{Metadata: map[string]string{"a": "100"}},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&TestCaseCommentLine{"  # a", nil},
			},
			TestFile{
				Tests: []TestCase{
					TestCase{Comments: []string{"  # a"}},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&CommandLine{"  $ a", nil},
			},
			TestFile{
				Tests: []TestCase{
					TestCase{
						Steps: []*TestStep{
							&TestStep{Command: "a"},
						},
					},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&CommandLine{"  $ a", nil},
				&TestCaseLine{"bbbb:"},
				&CommandLine{"  $ b", nil},
			},
			TestFile{
				Tests: []TestCase{
					TestCase{
						Steps: []*TestStep{
							&TestStep{Command: "a"},
						},
					},
					TestCase{
						Steps: []*TestStep{
							&TestStep{Command: "b"},
						},
					},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&CommandLine{"  $ a", nil},
				&OutputLine{"  a", nil},
			},
			TestFile{
				Tests: []TestCase{
					TestCase{
						Steps: []*TestStep{
							&TestStep{Command: "a", Output: "a"},
						},
					},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&CommandLine{"  $ a", nil},
				&OutputLine{"  a", nil},
				&CommandLine{"  $ b", nil},
				&CommandContinueLine{"  > c", nil},
			},
			TestFile{
				Tests: []TestCase{
					TestCase{
						Steps: []*TestStep{
							&TestStep{Command: "a", Output: "a"},
							&TestStep{Command: "b\nc"},
						},
					},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&CommandLine{"  $ a", nil},
				&OutputLine{"  a", nil},
				&CommandLine{"  $ b", nil},
				&CommandContinueLine{"  > c", nil},
				&OutputLine{"  b", nil},
				&OutputLine{"  c", nil},
			},
			TestFile{
				Tests: []TestCase{
					TestCase{
						Steps: []*TestStep{
							&TestStep{Command: "a", Output: "a"},
							&TestStep{Command: "b\nc", Output: "b\nc"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		testFile := NewFromRawT(tt.l)

		assert.Equal(t, tt.t, testFile)
	}
}
