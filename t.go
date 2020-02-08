package main

import (
	"bufio"
	"go.uber.org/zap"
	"io"
	"regexp"
)

var (
	CommentLineRegex         = regexp.MustCompile(`^\s*#.*$`)
	EmptyLineRegex           = regexp.MustCompile(`^\s*$`)
	MetaCommentLineRegex     = regexp.MustCompile(`^\s*# meta.*$`)
	TestCaseLineRegex        = regexp.MustCompile(`^.*:\s*$`)
	CommandLineRegex         = regexp.MustCompile(`^  \$ .*$`)
	CommandContinueLineRegex = regexp.MustCompile(`^  > .*$`)
	OutputLineRegex          = regexp.MustCompile(`^  [^>$].*$`)
)

type ParseTContext struct {
	t               []Line
	metaCommentLine *MetaCommentLine
	testCaseLine    *TestCaseLine
	commandLine     *CommandLine
}

func (c ParseTContext) isContext(context string) bool {
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

func parseCommentLine(line string, context *ParseTContext) error {
	c := CommentLine{line}
	context.t = append(context.t, &c)
	return nil
}

func parseMetaCommentLine(line string, context *ParseTContext) error {
	c := MetaCommentLine{line}

	context.t = append(context.t, &c)
	context.metaCommentLine = &c
	return nil
}

func parseTestCaseLine(line string, context *ParseTContext) error {
	c := TestCaseLine{line}
	context.t = append(context.t, &c)
	context.testCaseLine = &c
	context.metaCommentLine = nil
	return nil
}

func parseCommandLine(line string, context *ParseTContext) error {
	c := CommandLine{line}
	context.t = append(context.t, &c)
	context.commandLine = &c
	return nil
}

func parseOutputLine(line string, context *ParseTContext) error {
	c := OutputLine{line}
	context.t = append(context.t, &c)
	return nil
}

func parseCommandContinueLine(line string, context *ParseTContext) error {
	c := CommandContinueLine{line}
	context.t = append(context.t, &c)
	return nil
}

func parseEmptyLine(line string, context *ParseTContext) error {
	c := EmptyLine{line}
	context.t = append(context.t, &c)
	context.metaCommentLine = nil
	context.commandLine = nil
	return nil
}

func ParseT(stream io.Reader) ([]Line, error) {
	scanner := bufio.NewScanner(stream)

	context := ParseTContext{t: []Line{}}
	parseHandler := []struct {
		contextCondition string
		lineCondition    func(string) bool
		f                func(string, *ParseTContext) error
	}{
		{"command", OutputLineRegex.MatchString, parseOutputLine},
		{"command", CommandContinueLineRegex.MatchString, parseCommandContinueLine},
		{"test-case", CommandLineRegex.MatchString, parseCommandLine},
		{"meta-comment", CommentLineRegex.MatchString, parseMetaCommentLine},
		{"", EmptyLineRegex.MatchString, parseEmptyLine},
		{"", MetaCommentLineRegex.MatchString, parseMetaCommentLine},
		{"", CommentLineRegex.MatchString, parseCommentLine},
		{"", TestCaseLineRegex.MatchString, parseTestCaseLine},
	}

	for scanner.Scan() {
		line := scanner.Text()
		zap.S().Debug(line)

		// TODO: add error return. (ex. meta w/o test-case)
		handled := false
		for idx, handler := range parseHandler {
			okContext := context.isContext(handler.contextCondition)
			okLine := handler.lineCondition(line)
			zap.S().Debug("Handler#", idx, okContext, okLine)
			if okContext && okLine {
				handler.f(line, &context)
				handled = true
				break
			}
		}

		if !handled {
			zap.S().Warn("line not processed")
		}
	}
	return context.t, nil
}
