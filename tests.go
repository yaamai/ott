package main

import (
	"go.uber.org/zap"
	"regexp"
	"strings"
)

type TestFile struct {
	Name     string            `json:"name"`
	Comments []string          `json:"comments"`
	Metadata map[string]string `json:"metadata"`
	Tests    []*TestCase       `json:"tests"`
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

// Convert "raw-T-file" to structual representation
// this convert is irreversible
func NewFromRawT(name string, rawT []Line) TestFile {
	// TODO: refactoring

	var (
		t        TestFile
		meta     *map[string]string
		comments []string
		testCase *TestCase
		testStep *TestStep
	)

	t.Name = name

	for _, rawLine := range rawT {
		switch line := rawLine.(type) {
		case *CommentLine:
			zap.S().Debug("CommentLine:", line.Line())
			comments = append(comments, line.Line())
		case *EmptyLine:
			zap.S().Debug("EmptyLine:", line.Line())
			// to first empty-line, comments belong to file
			if t.Comments == nil {
				t.Comments = comments
				comments = []string{}
			}
			if meta != nil && testCase == nil {
				t.Metadata = *meta
				meta = nil
			}
		case *MetaCommentLine:
			zap.S().Debug("MetaCommentLine:", line.Line())
			if meta == nil {
				meta = &map[string]string{}
				// skip meta start-marker
				continue
			}
			match := ParseMetaCommentLine.FindStringSubmatch(line.Line())
			zap.S().Debug("ParseMetaCommentLine:", line.Line(), match)

			key := match[1]
			value := match[2]
			(*meta)[key] = value
		case *TestCaseLine:
			zap.S().Debug("TestCaseLine:", line.Line())
			if testCase != nil {
				t.Tests = append(t.Tests, testCase)
			}
			match := ParseTestCaseLine.FindStringSubmatch(line.Line())
			testCase = &TestCase{Name: match[1]}
			if meta != nil {
				testCase.Metadata = *meta
				meta = nil
			}
			if len(comments) != 0 {
				testCase.Comments = comments
				comments = []string{}
			}
		case *CommandLine:
			zap.S().Debug("CommandLine:", line.Line())
			testStep = &TestStep{}
			testCase.Steps = append(testCase.Steps, testStep)
			match := ParseCommandLine.FindStringSubmatch(line.Line())
			testStep.Command = match[1]
			if len(comments) != 0 {
				testStep.Comments = comments
				comments = []string{}
			}
		case *OutputLine:
			zap.S().Debug("OutputLine:", line.Line())

			match := ParseOutputLine.FindStringSubmatch(line.Line())
			testStep.ExpectedOutput += match[1] + "\n"
		case *CommandContinueLine:
			zap.S().Debug("CommandContinueLine:", line.Line())
			match := ParseCommandLine.FindStringSubmatch(line.Line())
			testStep.Command += "\n" + match[1]
		}
	}
	if testCase != nil {
		t.Tests = append(t.Tests, testCase)
	}

	return t
}

func (t *TestFile) ConvertToLines(mode string) []Line {
	lines := []Line{}

	// add file comments
	for _, c := range t.Comments {
		lines = append(lines, &CommentLine{c})
	}

	for _, testCase := range t.Tests {
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

			if mode == "diff" {
				// add diff
				if testStep.Diff != "" {
					for _, l := range strings.Split(testStep.Diff, "\n") {
						lines = append(lines, &OutputLine{"  " + l})
					}
				}
			} else if mode == "actual" {
				// add actual output
				if testStep.ActualOutput != "" {
					for _, l := range strings.Split(testStep.ActualOutput, "\n") {
						lines = append(lines, &OutputLine{"  " + l})
					}
				}
			} else {
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
