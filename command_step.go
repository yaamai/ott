package main

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// Code represents runnable in shell
type Code interface {
	Run(s *ShellSession) CodeResult
	StringLines() []string
}

// TemplateCode represents template runnable in shell
type TemplateCode interface {
	RunTemplate(s *ShellSession) []string
	StringLines() []string
}

// CodeResult represents checkable with shell
type CodeResult interface {
	Check() bool
	StringLines() []string
}

// RunnerHook provide hook interface when running
type RunnerHook interface {
	onFileStart(filename string)
	onFileEnd(filename string, fileBytes []byte, outputBytes []byte, result []CodeResult)
	onCodeBlockStart(cbName string)
	onCodeBlockEnd(cbName string)
	onCodeStart(cbName string, c Code)
	onCodeEnd(cbName string, r CodeResult)
}

// RunMarkdown run a markdown bytes
func RunMarkdown(filename string, bytes []byte, sess *ShellSession, hook RunnerHook) []CodeResult {
	hook.onFileStart(filename)

	fileResults := []CodeResult{}
	_, modified := walkCodeBlocks(bytes, func(n ast.Node, lines []string) []string {
		name := getNearestHeading(bytes, n)
		hook.onCodeBlockStart(name)

		cbResults := []CodeResult{}
		walkCodes(lines, sess, func(c Code) {
			hook.onCodeStart(name, c)
			r := c.Run(sess)
			hook.onCodeEnd(name, r)
			cbResults = append(cbResults, r)
		})
		fileResults = append(fileResults, cbResults...)

		hook.onCodeBlockEnd(name)
		return convertCommandStepResults(cbResults)
	})

	hook.onFileEnd(filename, bytes, modified, fileResults)

	return fileResults
}

func parseCode(kind string, buf map[string][]string) interface{} {
	switch kind {
	case "<< ":
		return &TemplateCommand{TemplateCommand: append(buf["<< "], buf["> "]...)}
	case "$ ":
		out, chk := parseOutput(buf[""])
		return &Command{Command: append(buf["$ "], buf["> "]...), Output: out, Checker: chk}
	case "# ":
		out, chk := parseOutput(buf[""])
		return &Command{Command: append(buf["# "], buf["> "]...), Output: out, Checker: chk}
	default:
		return nil
	}
}

func walkAnyCodes(lines []string, f func(c interface{})) {
	codeSeparater := []string{"<< ", "$ ", "# "}
	codePrefixes := []string{"> "}

	codeKind := ""
	buf := map[string][]string{}
	for _, l := range lines {
		p := hasAnyPrefix(l, codeSeparater...)
		if p == "" {
			p = hasAnyPrefix(l, codePrefixes...)
			buf[p] = append(buf[p], strings.TrimPrefix(l, p))
			continue
		}

		if codeKind != "" {
			f(parseCode(codeKind, buf))
			buf = map[string][]string{}
		}
		codeKind = p
		buf[p] = append(buf[p], strings.TrimPrefix(l, p))
	}
	if codeKind != "" && len(buf) > 0 {
		f(parseCode(codeKind, buf))
	}
}

func walkCodes(lines []string, sess *ShellSession, f func(c Code)) {
	var expandTemplate func(c interface{})
	expandTemplate = func(c interface{}) {
		switch v := c.(type) {
		case TemplateCode:
			walkAnyCodes(v.RunTemplate(sess), expandTemplate)
		case Code:
			f(v)
		}
	}
	walkAnyCodes(lines, expandTemplate)
}

// TemplateCommand represents a template command-line
type TemplateCommand struct {
	TemplateCommand []string `json:"template_command"`
	Template        []string
}

// RunTemplate execute TemplateCommand and return CodeResult
func (c *TemplateCommand) RunTemplate(s *ShellSession) []string {
	_, result := s.Run(strings.Join(c.TemplateCommand, "\n") + "\n")
	c.Template = strings.Split(result, "\n")
	return c.Template
}

// Command represents a command-line
type Command struct {
	Command []string         `json:"command"`
	Output  []string         `json:"output"`
	Checker []CommandChecker `json:"checker"`
}

// CommandResult represents executed CommandStep results
type CommandResult struct {
	Command
	ActualOutput []string `json:"actual"`
	Rc           int      `json:"rc"`
}

func parseOutput(lines []string) ([]string, []CommandChecker) {
	output := []string{}
	checker := []CommandChecker{}
	for _, l := range lines {
		if m := NewRcChecker(l); m != nil {
			checker = append(checker, m)
		} else if m := NewHasChecker(l); m != nil {
			checker = append(checker, m)
		} else {
			output = append(output, l)
		}
	}
	for idx := len(output)-1; idx >= 0; idx-- {
		if len(output[idx]) != 0 {
			output = output[0:idx+1]
			break
		}
	}

	return output, checker
}

// Run execute CommandStep and return CodeResult
func (c Command) Run(s *ShellSession) CodeResult {
	rc, result := s.Run(strings.Join(c.Command, "\n") + "\n")
	o := CommandResult{Command: c, ActualOutput: strings.Split(result, "\n"), Rc: rc}
	return o
}

// Check checks CommandStepResult is expected outputs or not
func (c CommandResult) Check() bool {
	// check special matcher
	for idx := range c.Checker {
		if !c.Checker[idx].IsMatch(c) {
			return false
		}
	}

	if len(c.Output) > 0 && len(c.Output) != len(c.ActualOutput) {
		return false
	}

	for idx := range c.Output {
		if c.Output[idx] != c.ActualOutput[idx] {
			return false
		}
	}

	return true
}

// StringLines convert CommandStepResult to array of string
func (c Command) StringLines() []string {
	result := []string{}
	prompt := "#"
	for _, cmd := range c.Command {
		result = append(result, fmt.Sprintf("%s %s\n", prompt, cmd))
		prompt = ">"
	}

	return result
}

// StringLines convert CommandStepResult to array of string
func (c CommandResult) StringLines() []string {
	result := []string{}
	prompt := "#"
	for _, cmd := range c.Command.Command {
		result = append(result, fmt.Sprintf("%s %s\n", prompt, cmd))
		prompt = ">"
	}
	for _, out := range c.ActualOutput {
		result = append(result, fmt.Sprintf("%s\n", out))
	}

	return result
}

// StringLines convert CommandStepResult to array of string
func (c TemplateCommand) StringLines() []string {
	return nil
	// return c.Command.StringLines()
}

func countCommandStepResults(results []CodeResult) (string, int, int) {
	success := 0
	fail := 0
	for _, step := range results {
		if step.Check() {
			success++
		} else {
			fail++
		}
	}

	s := "OK"
	if fail > 0 {
		s = "FAIL"
	}
	return s, success, fail
}

func convertCommandStepResults(results []CodeResult) []string {
	result := []string{}
	for _, r := range results {
		result = append(result, r.StringLines()...)
	}
	return result
}
