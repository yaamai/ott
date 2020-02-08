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
			[]Line{&MetaCommentLine{"# meta:"}},
			TestFile{},
		},
		{
			[]Line{
				&MetaCommentLine{"# meta:"},
				&MetaCommentLine{"#  a: 100"},
			},
			TestFile{},
		},
		{
			[]Line{
				&MetaCommentLine{"# meta:"},
				&MetaCommentLine{"#  a: 100"},
				&TestCaseLine{"aaaa:"},
			},
			TestFile{
				Tests: []*TestCase{
					&TestCase{Name: "aaaa", Metadata: map[string]string{"a": "100"}},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&TestCaseCommentLine{"  # a"},
			},
			TestFile{
				Tests: []*TestCase{
					&TestCase{Name: "aaaa", Comments: []string{"  # a"}},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&CommandLine{"  $ a"},
			},
			TestFile{
				Tests: []*TestCase{
					&TestCase{
						Name: "aaaa",
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
				&CommandLine{"  $ a"},
				&TestCaseLine{"bbbb:"},
				&CommandLine{"  $ b"},
			},
			TestFile{
				Tests: []*TestCase{
					&TestCase{
						Name: "aaaa",
						Steps: []*TestStep{
							&TestStep{Command: "a"},
						},
					},
					&TestCase{
						Name: "bbbb",
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
				&CommandLine{"  $ a"},
				&OutputLine{"  a"},
			},
			TestFile{
				Tests: []*TestCase{
					&TestCase{
						Name: "aaaa",
						Steps: []*TestStep{
							&TestStep{Command: "a", ExpectedOutput: "a\n"},
						},
					},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&CommandLine{"  $ a"},
				&OutputLine{"  a"},
				&CommandLine{"  $ b"},
				&CommandContinueLine{"  > c"},
			},
			TestFile{
				Tests: []*TestCase{
					&TestCase{
						Name: "aaaa",
						Steps: []*TestStep{
							&TestStep{Command: "a", ExpectedOutput: "a\n"},
							&TestStep{Command: "b\nc"},
						},
					},
				},
			},
		},
		{
			[]Line{
				&TestCaseLine{"aaaa:"},
				&CommandLine{"  $ a"},
				&OutputLine{"  a"},
				&CommandLine{"  $ b"},
				&CommandContinueLine{"  > c"},
				&OutputLine{"  b"},
				&OutputLine{"  c"},
			},
			TestFile{
				Tests: []*TestCase{
					&TestCase{
						Name: "aaaa",
						Steps: []*TestStep{
							&TestStep{Command: "a", ExpectedOutput: "a\n"},
							&TestStep{Command: "b\nc", ExpectedOutput: "b\nc\n"},
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
