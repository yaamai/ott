package main

import (
	"bufio"
	"go.uber.org/zap"
	"io"
	"regexp"
)

var (
	CommentLineRegex          = regexp.MustCompile(`^\s*#.*$`)
	MetaCommentLineRegex      = regexp.MustCompile(`^\s*# meta.*$`)
	TestCaseLineRegex         = regexp.MustCompile(`^.*:\s*$`)
	CommandLineRegex          = regexp.MustCompile(`^  \$ .*$`)
	CommandContinueLineRegex  = regexp.MustCompile(`^  > .*$`)
	OutputLineRegex           = regexp.MustCompile(`^  [^>$].*$`)
)


type Line interface {
	Type() string
	Line() string
    Equal(l Line) bool
}

type CommentLine struct {
    string
}
func (c *CommentLine) Type() string {
    return "comment"
}
func (c *CommentLine) Line() string {
    return c.string
}
func (c *CommentLine) Equal(l Line) bool {
    return c.Line() == l.Line()
}

type MetaCommentLine struct {
    string
    parent *MetaCommentLine
}
func (c *MetaCommentLine) Type() string {
    return "meta-comment"
}
func (c *MetaCommentLine) Line() string {
    return c.string
}
func (c *MetaCommentLine) Equal(l Line) bool {
    return c.Line() == l.Line()
}

type TestCaseLine struct {
    string
}
func (c *TestCaseLine) Type() string {
    return "test-case"
}
func (c *TestCaseLine) Line() string {
    return c.string
}
func (c *TestCaseLine) Equal(l Line) bool {
    return c.Line() == l.Line()
}

type TestCaseCommentLine struct {
    string
    parent *TestCaseLine
}
func (c *TestCaseCommentLine) Type() string {
    return "test-case-comment"
}
func (c *TestCaseCommentLine) Line() string {
    return c.string
}
func (c *TestCaseCommentLine) Equal(l Line) bool {
    return c.Line() == l.Line()
}

type CommandLine struct {
    string
    parent *TestCaseLine
}
func (c *CommandLine) Type() string {
    return "command"
}
func (c *CommandLine) Line() string {
    return c.string
}
func (c *CommandLine) Equal(l Line) bool {
    return c.Line() == l.Line()
}

type OutputLine struct {
    string
    parent *CommandLine
}
func (c *OutputLine) Type() string {
    return "output"
}
func (c *OutputLine) Line() string {
    return c.string
}
func (c *OutputLine) Equal(l Line) bool {
    return c.Line() == l.Line()
}

type CommandContinueLine struct {
    string
    parent *CommandLine
}
func (c *CommandContinueLine) Type() string {
    return "output"
}
func (c *CommandContinueLine) Line() string {
    return c.string
}
func (c *CommandContinueLine) Equal(l Line) bool {
    return c.Line() == l.Line()
}

type ParseRawTContext struct {
	t []Line
    metaCommentLine *MetaCommentLine
    testCaseLine *TestCaseLine
    commandLine *CommandLine
}
func (c ParseRawTContext) isContext(context string) bool {
	switch context {
	case "":
		return true
    case "meta-comment":
		return c.metaCommentLine != nil
    case "test-case":
		return c.testCaseLine != nil
    case "command":
		return c.commandLine != nil
	default:
		return false
	}
}

func parseCommentLine(line string, context *ParseRawTContext) error {
    c := CommentLine{line}
    context.t = append(context.t, &c)
	return nil
}

func parseMetaCommentLine(line string, context *ParseRawTContext) error {
    c := MetaCommentLine{line, nil}

    if context.metaCommentLine != nil {
        c.parent = context.metaCommentLine
    }
    context.t = append(context.t, &c)
    context.metaCommentLine = &c
	return nil
}

func parseTestCaseLine(line string, context *ParseRawTContext) error {
    c := TestCaseLine{line}
    context.t = append(context.t, &c)
    context.testCaseLine = &c
	return nil
}

func parseTestCaseCommentLine(line string, context *ParseRawTContext) error {
    c := TestCaseCommentLine{line, context.testCaseLine}
    context.t = append(context.t, &c)
	return nil
}

func parseCommandLine(line string, context *ParseRawTContext) error {
    c := CommandLine{line, context.testCaseLine}
    context.t = append(context.t, &c)
    context.commandLine = &c
	return nil
}

func parseOutputLine(line string, context *ParseRawTContext) error {
    c := OutputLine{line, context.commandLine}
    context.t = append(context.t, &c)
	return nil
}

func parseCommandContinueLine(line string, context *ParseRawTContext) error {
    c := CommandContinueLine{line, context.commandLine}
    context.t = append(context.t, &c)
	return nil
}

func ParseRawT(stream io.Reader) ([]Line, error) {
	scanner := bufio.NewScanner(stream)

	context := ParseRawTContext{t: []Line{}}
	parseHandler := []struct {
		contextCondition string
		lineCondition    func(string) bool
		f                func(string, *ParseRawTContext) error
	}{
		{"command", OutputLineRegex.MatchString, parseOutputLine},
		{"command", CommandContinueLineRegex.MatchString, parseCommandContinueLine},
		{"test-case", CommandLineRegex.MatchString, parseCommandLine},
		{"test-case", CommentLineRegex.MatchString, parseTestCaseCommentLine},
		{"meta-comment", CommentLineRegex.MatchString, parseMetaCommentLine},
		{"", MetaCommentLineRegex.MatchString, parseMetaCommentLine},
		{"", CommentLineRegex.MatchString, parseCommentLine},
		{"", TestCaseLineRegex.MatchString, parseTestCaseLine},
	}

	for scanner.Scan() {
		line := scanner.Text()
		zap.S().Debug(line)

		// TODO: add error return. (ex. meta w/o test-case)
		for idx, handler := range parseHandler {
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
