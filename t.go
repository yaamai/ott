package main

import (
	"bufio"
	"encoding/json"
	"io"
	"regexp"
    "go.uber.org/zap"
    "strings"
)

type Lineable interface {
	Type() string
	Lines() []string
}

type TFile struct {
	Lines []Lineable
}

var (
	CommentLineRe          = regexp.MustCompile(`^\s*#.*$`)
	MetaCommentLineRe      = regexp.MustCompile(`^\s*# meta.*$`)
	MetaDataLineRe         = regexp.MustCompile(`^\s*#\s+(.*?):\s*(.*?)$`)
	TestCaseLineRe         = regexp.MustCompile(`^.*:\s*$`)
	EmptyTestStepLineRe         = regexp.MustCompile(`^(  |  # .*)$`)
	TestStepLineRe         = regexp.MustCompile(`^  \$ .*$`)
	TestStepContinueLineRe = regexp.MustCompile(`^  > .*$`)
	TestStepOutputLineRe   = regexp.MustCompile(`^  [^>$].*$`)
    ParseTestStepCommand = regexp.MustCompile(`^  [>$]\s*(.*)$`)
    ParseTestStepOutput = regexp.MustCompile(`^  (.*)$`)
)

type Context struct{
	t *TFile
	testCase *TestCase
	testStep *TestStep
	testCaseMeta *TestMeta
}

func (c Context) isContext(context string) bool {
	switch context {
	case "":
		return true
	case "testcase-meta":
		return c.testCaseMeta != nil
	case "testcase":
		return c.testCase != nil
	case "teststep":
		return c.testCase != nil && c.testStep != nil
	default:
		return true
	}
}

func parseTestCaseMetaStart(line string, context *Context) error {
	context.testCaseMeta = &TestMeta{String: line, Meta: map[string]string{}}
	return nil
}

func parseTestCaseMeta(line string, context *Context) error {
	match := MetaDataLineRe.FindStringSubmatch(line)
	key := match[1]
	value := match[2]
	context.testCaseMeta.Meta[key] = value
	return nil
}

func parseComment(line string, context *Context) error {
	c := Comment{line}
	context.t.Lines = append(context.t.Lines, &c)
	return nil
}

func parseTestCaseStart(line string, context *Context) error {
	testCase := TestCase{Name: line}
	if context.testCaseMeta != nil {
		testCase.Metadata = *context.testCaseMeta
		context.testCaseMeta = nil
	}
	context.t.Lines = append(context.t.Lines, &testCase)
	context.testCase = &testCase
	return nil
}

func parseTestStep(line string, context *Context) error {
	testStep := TestStep{Commands: []string{line}}
	context.testCase.TestSteps = append(context.testCase.TestSteps, &testStep)
	context.testStep = &testStep
	return nil
}

func parseTestContinueStep(line string, context *Context) error {
	context.testStep.Commands = append(context.testStep.Commands, line)
	return nil
}

func parseTestStepOutput(line string, context *Context) error {
	context.testStep.Output = append(context.testStep.Output, line)
	return nil
}

func parseEmptyCommentTestStep(line string, context *Context) error {
	testStep := TestStep{EmptyString: line}
	context.testCase.TestSteps = append(context.testCase.TestSteps, &testStep)
	context.testStep = nil
	return nil
}

func ParseTFile(stream io.Reader) (*TFile, error) {
	scanner := bufio.NewScanner(stream)

	context := Context{t: &TFile{}}
	parseHandler := []struct{
		contextCondition string
		lineCondition func(string)bool
		f func(string, *Context)error
	}{
		{"", MetaCommentLineRe.MatchString, parseTestCaseMetaStart},
		{"testcase-meta", CommentLineRe.MatchString, parseTestCaseMeta},
		{"testcase", EmptyTestStepLineRe.MatchString, parseEmptyCommentTestStep},
		{"", CommentLineRe.MatchString, parseComment},
		{"", TestCaseLineRe.MatchString, parseTestCaseStart},
		{"testcase", TestStepLineRe.MatchString, parseTestStep},
		{"teststep", TestStepContinueLineRe.MatchString, parseTestContinueStep},
		{"teststep", TestStepOutputLineRe.MatchString, parseTestStepOutput},

	}

	for scanner.Scan() {
		line := scanner.Text()
        zap.S().Debug(line)

		// TODO: add error return. (ex. meta w/o test-case)
		for idx, handler := range(parseHandler)  {
			okContext := context.isContext(handler.contextCondition)
			okLine := handler.lineCondition(line)
            zap.S().Debug("Handler#", idx, okContext, okLine)
			if okContext && okLine {
				handler.f(line, &context)
				break
			}
		}
	}
	return context.t, nil
}

func (t *TFile) UnmarshalJSON(b []byte) error {
	// parse json array only
	m := []json.RawMessage{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	// detect "type" value
	typeStruct := struct{ Type string }{}
	for _, line := range m {
		err := json.Unmarshal(line, &typeStruct)
		if err != nil {
			return err
		}

		var dst interface{}
		switch typeStruct.Type {
		case "comment":
			dst = new(Comment)
		case "testcase":
			dst = new(TestCase)
		}

		// unmarshal to dedicated type
		err = json.Unmarshal(line, dst)
		if err != nil {
			return err
		}
		t.Lines = append(t.Lines, dst.(Lineable))
	}

	return nil
}

type Comment struct {
	String string `json:"string"`
}

func (c *Comment) Type() string {
	return "comment"
}
func (c *Comment) Lines() []string {
	return []string{c.String}
}

type TestCase struct {
	Metadata  TestMeta    `json:"metadata"`
	Name      string      `json:"name"`
	TestSteps []*TestStep `json:"steps"`
}

func (t *TestCase) Type() string {
	return "testcase"
}
func (t *TestCase) Lines() []string {
	return []string{t.Name}
}

type TestMeta struct {
	String string            `json:"string"`
	Meta   map[string]string `json:"meta"`
}

type TestStep struct {
    EmptyString string `json:"empty_string"`
	Commands []string `json:"commands"`
	Output   []string `json:"outputs"`
}

func (t *TestStep) GetCommand() string {
    builder := new(strings.Builder)
    for _, command := range(t.Commands) {
        matches := ParseTestStepCommand.FindStringSubmatch(command)
        builder.WriteString(matches[1])
    }

    return builder.String()
}

func (t *TestStep) GetOutput() string {
    builder := new(strings.Builder)
    for _, output := range(t.Output) {
        matches := ParseTestStepOutput.FindStringSubmatch(output)
        builder.WriteString(matches[1])
        builder.WriteString("\r\n")
    }

    return builder.String()
}
