package main

import (
	"regexp"
)

type TestFile struct {
    Name string `json:"name"`
    Comments []string `json:"comments"`
    Tests []TestCase `json:"tests"`
}

type TestCase struct {
    Name string `json:"name"`
    Metadata map[string]string `json:"metadata"`
    Comments []string `json:"comments"`
    Steps []TestStep `json:"steps"`
}

type TestStep struct {
    Comments []string `json:"comments"`
    Command string `json:"command"`
    Output string `json:"Output"`
}

var (
	ParseMetaCommentLine = regexp.MustCompile(`^\s*#\s+(.*?):\s*(.*?)$`)
)

// Convert "raw-T-file" to structual representation
// this convert is irreversible
func NewFromRawT(rawT []Line) TestFile {
    // TODO: name from file-meta or filename or parameter?

    var (
        t TestFile
        meta *map[string]string
        testCase *TestCase
        testStep *TestStep
    )

    for _, rawLine := range(rawT) {
        switch line := rawLine.(type) {
        case *CommentLine:
            t.Comments = append(t.Comments, line.Line())
        case *MetaCommentLine:
            if meta == nil {
                meta = &map[string]string{}
            }
	        match := ParseMetaCommentLine.FindStringSubmatch(line.Line())
	        key := match[1]
	        value := match[2]
            (*meta)[key] = value
        case *TestCaseLine:
            if testCase == nil {
                testCase = &TestCase{}
            }
        case *TestCaseCommentLine:
            testCase.Comments = append(testCase.Comments, line.Line())
        case *CommandLine:
            if testStep == nil {
                testStep = &TestStep{}
            }
            testStep.Command = line.Line()
        case *OutputLine:
            testStep.Output = line.Line()
        case *CommandContinueLine:
            testStep.Command = line.Line()
        }
    }

    return t
}

func (t *TestFile) ConvertToLines() []Line {
    return nil
}
