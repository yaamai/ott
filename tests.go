package main

import (
	"go.uber.org/zap"
	"regexp"
	"strings"
)

type TestFile struct {
	Name      string            `json:"name"`
	Comments  []string          `json:"comments"`
	Metadata  map[string]string `json:"metadata"`
	Tests     []*TestCase       `json:"tests"`
	Generated bool              `json:"generated"`
	NoInject  bool              `json:"no_inject"`
}

type TestCase struct {
	Name      string            `json:"name"`
	Metadata  map[string]string `json:"metadata"`
	Comments  []string          `json:"comments"`
	Steps     []*TestStep       `json:"steps"`
	Generated bool              `json:"generated"`
	NoInject  bool              `json:"no_inject"`
}

type TestStep struct {
	Comments       []string `json:"comments"`
	Command        string   `json:"command"`
	ExpectedOutput string   `json:"expected_output"`
	ActualOutput   string   `json:"actual_output"`
	Diff           string   `json:"diff"`
}

var (
	ParseMetaCommentLine = regexp.MustCompile(`^\s*#\s+(.*?):\s*(.*?)$`)
	ParseCommandLine     = regexp.MustCompile(`^  [>$] (.*)$`)
	ParseOutputLine      = regexp.MustCompile(`^  (.*)$`)
	ParseTestCaseLine    = regexp.MustCompile(`^(.*):\s*$`)
)

func handleCommentLine(context *ParseTestsContext, c *CommentLine) {
	zap.S().Debug("CommentLine:", c.Line())
	context.comments = append(context.comments, c.Line())
}

func handleEmptyLine(context *ParseTestsContext, c *EmptyLine) {
	zap.S().Debug("EmptyLine:", c.Line())

	// to first empty-line, comments belong to file
	if context.t.Comments == nil {
		context.t.Comments = context.comments
		context.comments = []string{}
	}
	if context.meta != nil && context.testCase == nil {
		context.t.Metadata = *context.meta
		context.meta = nil
	}
}

func handleMetaCommentLine(context *ParseTestsContext, c *MetaCommentLine) {
	zap.S().Debug("MetaCommentLine:", c.Line())

	if context.meta == nil {
		context.meta = &map[string]string{}
		// skip meta start-marker
		return
	}
	match := ParseMetaCommentLine.FindStringSubmatch(c.Line())
	zap.S().Debug("ParseMetaCommentLine:", c.Line(), match)

	key := match[1]
	value := match[2]
	(*context.meta)[key] = value
}

func handleTestCaseLine(context *ParseTestsContext, c *TestCaseLine) {
	zap.S().Debug("TestCaseLine:", c.Line())

	if context.testCase != nil {
		context.t.Tests = append(context.t.Tests, context.testCase)
	}

	match := ParseTestCaseLine.FindStringSubmatch(c.Line())
	context.testCase = &TestCase{Name: match[1]}
	if context.meta != nil {
		context.testCase.Metadata = *context.meta
		context.meta = nil
	}
	if len(context.comments) != 0 {
		context.testCase.Comments = context.comments
		context.comments = []string{}
	}
}

func handleCommandLine(context *ParseTestsContext, c *CommandLine) {
	zap.S().Debug("CommandLine:", c.Line())

	context.testStep = &TestStep{}
	context.testCase.Steps = append(context.testCase.Steps, context.testStep)
	match := ParseCommandLine.FindStringSubmatch(c.Line())
	context.testStep.Command = match[1]
	if len(context.comments) != 0 {
		context.testStep.Comments = context.comments
		context.comments = []string{}
	}
}

func handleOutputLine(context *ParseTestsContext, c *OutputLine) {
	zap.S().Debug("OutputLine:", c.Line())

	match := ParseOutputLine.FindStringSubmatch(c.Line())
	context.testStep.ExpectedOutput += match[1] + "\n"
}

func handleCommandContinueLine(context *ParseTestsContext, c *CommandContinueLine) {
	zap.S().Debug("CommandContinueLine:", c.Line())

	match := ParseCommandLine.FindStringSubmatch(c.Line())
	context.testStep.Command += "\n" + match[1]
}

func callHandler(context *ParseTestsContext, rawLine Line) {
	switch line := rawLine.(type) {
	case *CommentLine:
		handleCommentLine(context, line)
	case *EmptyLine:
		handleEmptyLine(context, line)
	case *MetaCommentLine:
		handleMetaCommentLine(context, line)
	case *TestCaseLine:
		handleTestCaseLine(context, line)
	case *CommandLine:
		handleCommandLine(context, line)
	case *OutputLine:
		handleOutputLine(context, line)
	case *CommandContinueLine:
		handleCommandContinueLine(context, line)
	}
}

type ParseTestsContext struct {
	t        TestFile
	meta     *map[string]string
	comments []string
	testCase *TestCase
	testStep *TestStep
}

// Convert "raw-T-file" to structual representation
// this convert is irreversible
func NewFromRawT(name string, rawT []Line) TestFile {
	context := ParseTestsContext{}
	context.t.Name = name

	for _, rawLine := range rawT {
		callHandler(&context, rawLine)
	}
	if context.testCase != nil {
		context.t.Tests = append(context.t.Tests, context.testCase)
	}

	return context.t
}


func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}


func (t *TestFile) ConvertToLines(mode string) []Line {
	lines := []Line{}

	// add file comments
	for _, c := range t.Comments {
		lines = append(lines, &CommentLine{c})
	}

	// add test-file metadata
	if t.Metadata != nil {
		lines = append(lines, &MetaCommentLine{"# meta"})
		for k, v := range t.Metadata {
			lines = append(lines, &MetaCommentLine{"# " + k + ": " + v})
		}
		lines = append(lines, &EmptyLine{""})
	}

	for _, testCase := range t.Tests {
		// add test-case comment
		for _, c := range testCase.Comments {
			lines = append(lines, &CommentLine{c})
		}

		// add test-case metadata
		if testCase.Metadata != nil {
			lines = append(lines, &MetaCommentLine{"# meta"})
			for k, v := range testCase.Metadata {
				lines = append(lines, &MetaCommentLine{"# " + k + ": " + v})
			}
		}

		// add test-case line
		lines = append(lines, &TestCaseLine{testCase.Name + ":"})
		for _, testStep := range testCase.Steps {
			// add test-step comment
			for _, c := range testStep.Comments {
				lines = append(lines, &CommentLine{c})
			}

			// add test-command
			for idx, l := range strings.Split(testStep.Command, "\n") {
				if idx == 0 {
					lines = append(lines, &CommandLine{"  $ " + l})
				} else {
					lines = append(lines, &CommandLine{"  > " + l})
				}
			}

            modeList := strings.Split(mode, "+")
            if contains(modeList, "actual") {
				// add actual output
				if testStep.ActualOutput != "" {
					for _, l := range strings.Split(testStep.ActualOutput, "\n") {
						lines = append(lines, &OutputLine{"  " + l})
					}
				}
            }
            if contains(modeList, "diff") {
				// add diff
				if testStep.Diff != "" {
					for _, l := range strings.Split(testStep.Diff, "\n") {
						lines = append(lines, &OutputLine{"  " + l})
					}
				}
			}
            if contains(modeList, "expected") {
				// add output
				if testStep.ExpectedOutput != "" {
					for _, l := range strings.Split(testStep.ExpectedOutput, "\n") {
						lines = append(lines, &OutputLine{"  " + l})
					}
				}
			}
		}
	}
	return lines
}
