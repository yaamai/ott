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
        // TODO: warn not processed line
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
