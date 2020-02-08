package main

import (
	"regexp"
    "strings"
)

type TestFile struct {
	Name     string     `json:"name"`
	Comments []string   `json:"comments"`
	Tests    []TestCase `json:"tests"`
}

type TestCase struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata"`
	Comments []string          `json:"comments"`
	Steps    []*TestStep       `json:"steps"`
}

type TestStep struct {
	Comments []string `json:"comments"`
	Command  string   `json:"command"`
	ExpectedOutput   string   `json:"expected_output"`
    ActualOutput string `json:"actual_output"`
    Diff string `json:"diff"`
}

var (
	ParseMetaCommentLine = regexp.MustCompile(`^\s*#\s+(.*?):\s*(.*?)$`)
	ParseCommandLine     = regexp.MustCompile(`^  [>$] (.*)$`)
	ParseOutputLine      = regexp.MustCompile(`^  (.*)$`)
	ParseTestCaseLine    = regexp.MustCompile(`^(.*):\s*$`)
)

// Convert "raw-T-file" to structual representation
// this convert is irreversible
func NewFromRawT(rawT []Line) TestFile {
	// TODO: name from file-meta or filename or parameter?

	var (
		t        TestFile
		meta     *map[string]string
		testCase *TestCase
		testStep *TestStep
	)

	for _, rawLine := range rawT {
		switch line := rawLine.(type) {
		case *CommentLine:
			t.Comments = append(t.Comments, line.Line())
		case *MetaCommentLine:
			if meta == nil {
				meta = &map[string]string{}
				// skip meta start-marker
				continue
			}
			match := ParseMetaCommentLine.FindStringSubmatch(line.Line())
			key := match[1]
			value := match[2]
			(*meta)[key] = value
		case *TestCaseLine:
			if testCase != nil {
				t.Tests = append(t.Tests, *testCase)
			}
            match := ParseTestCaseLine.FindStringSubmatch(line.Line())
			testCase = &TestCase{Name: match[1]}
			if meta != nil {
				testCase.Metadata = *meta
				meta = nil
			}
		case *TestCaseCommentLine:
			testCase.Comments = append(testCase.Comments, line.Line())
		case *CommandLine:
			testStep = &TestStep{}
			testCase.Steps = append(testCase.Steps, testStep)
			match := ParseCommandLine.FindStringSubmatch(line.Line())
			testStep.Command = match[1]
		case *OutputLine:
			if testStep.ExpectedOutput != "" {
				testStep.ExpectedOutput += "\n"
			}

			match := ParseOutputLine.FindStringSubmatch(line.Line())
			testStep.ExpectedOutput += match[1]
		case *CommandContinueLine:
			match := ParseCommandLine.FindStringSubmatch(line.Line())
			testStep.Command += "\n" + match[1]
		}
	}
	if testCase != nil {
		t.Tests = append(t.Tests, *testCase)
	}

	return t
}

func (t *TestFile) ConvertToLines(mode string) []Line {
    lines := []Line{}

    // add file comments
    for _, c := range(t.Comments) {
        lines = append(lines, &CommentLine{c})
    }

	for _, testCase := range t.Tests {
        // add test-case metadata
        if testCase.Metadata != nil {
            lines = append(lines, &MetaCommentLine{"# meta"})
            for k, v := range(testCase.Metadata) {
                lines = append(lines, &MetaCommentLine{"# " + k + ": " + v})
            }
        }

        // add test-case line
        lines = append(lines, &TestCaseLine{testCase.Name + ":"})
		for _, testStep := range testCase.Steps {
            // add test-step comment
            for _, c := range(testStep.Comments) {
                lines = append(lines, &CommentLine{c})
            }

            // add test-command
            for idx, l := range(strings.Split(testStep.Command, "\n")) {
                if idx == 0 {
                    lines = append(lines, &CommandLine{"  $ " + l})
                } else {
                    lines = append(lines, &CommandLine{"  > " + l})
                }
            }

            if mode == "diff" {
                // add diff
                if testStep.Diff != "" {
                    for _, l := range(strings.Split(testStep.Diff, "\n")) {
                        lines = append(lines, &OutputLine{"  " + l})
                    }
                }
            } else if mode == "actual" {
                // add actual output
                if testStep.ActualOutput != "" {
                    for _, l := range(strings.Split(testStep.ActualOutput, "\n")) {
                        lines = append(lines, &OutputLine{"  " + l})
                    }
                }
            } else {
                // add output
                if testStep.ExpectedOutput != "" {
                    for _, l := range(strings.Split(testStep.ExpectedOutput, "\n")) {
                        lines = append(lines, &OutputLine{"  " + l})
                    }
                }
            }
        }
    }
	return lines
}
