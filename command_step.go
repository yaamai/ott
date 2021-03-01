package main

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
)

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
		cb := ParseCodeBlock(name, lines)
		hook.onCodeBlockStart(name)

		cbResults := []CodeResult{}
		for _, c := range cb.Codes {
			hook.onCodeStart(name, c)
			r := c.Run(sess)
			hook.onCodeEnd(name, r)
			cbResults = append(cbResults, r)
		}
		fileResults = append(fileResults, cbResults...)

		hook.onCodeBlockEnd(name)
		return convertCommandStepResults(cbResults)
	})

	hook.onFileEnd(filename, bytes, modified, fileResults)

	return fileResults
}

// RunCodeBlock run a codeblock and call hooks
func RunCodeBlock(lines []string, s *ShellSession, hooks RunnerHook) {
	result := []Code{}
	for len(lines) > 0 {
		rest, code := readLinesToFunc(lines, func(l string) (bool, string) {
			p := hasAnyPrefix(l, "<< ", "$ ", "# ")
			return p != "", l
		})
		lines = rest

		if c := newTemplateCommand(code); c != nil {
			result = append(result, c)
		} else if c := newCommand(code); c != nil {
			r := c.Run(s)
			result = append(result, c)
		} else {
			break
		}
	}
}

// CodeBlock represents a CodeBlock in markdown
type CodeBlock struct {
	Name  string `json:"name"`
	Codes []Code
}

// Code represents runnable in shell
type Code interface {
	Run(s *ShellSession) CodeResult
	StringLines() []string
}

// CodeResult represents checkable with shell
type CodeResult interface {
	Check() bool
	StringLines() []string
}

// TemplateCommand represents a meta command-line
type TemplateCommand struct {
	Codes           []Code
	TemplateCommand []string `json:"template_command"`
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

func newTemplateCommand(lines []string) *TemplateCommand {
	if hasAnyPrefix(lines[0], "<< ") == "" {
		return nil
	}
	_, command := readLinesToFunc(lines, func(l string) (bool, string) {
		p := hasAnyPrefix(l, "<< ", "> ")
		return p == "", strings.TrimPrefix(l, p)
	})

	return &TemplateCommand{TemplateCommand: command}
}

func readLinesToFunc(lines []string, f func(l string) (bool, string)) ([]string, []string) {
	r := []string{}
	for idx, l := range lines {
		end, l := f(l)
		if end && idx > 0 {
			return lines[idx:], r
		}
		r = append(r, l)
	}
	return []string{}, r
}

func newCommand(lines []string) *Command {
	if hasAnyPrefix(lines[0], "$ ", "# ") == "" {
		return nil
	}
	rest, command := readLinesToFunc(lines, func(l string) (bool, string) {
		p := hasAnyPrefix(l, "$ ", "# ", "> ")
		return p == "", strings.TrimPrefix(l, p)
	})

	output := []string{}
	checker := []CommandChecker{}
	for _, l := range rest {
		if m := NewRcChecker(l); m != nil {
			checker = append(checker, m)
		} else if m := NewHasChecker(l); m != nil {
			checker = append(checker, m)
		} else {
			output = append(output, l)
		}
	}

	return &Command{Command: command, Output: output, Checker: checker}
}

// ParseCodeBlock parses code block string lines to CodeBlock
func ParseCodeBlock(name string, lines []string) CodeBlock {
	result := []Code{}
	for len(lines) > 0 {
		rest, code := readLinesToFunc(lines, func(l string) (bool, string) {
			p := hasAnyPrefix(l, "<< ", "$ ", "# ")
			return p != "", l
		})
		lines = rest

		if c := newTemplateCommand(code); c != nil {
			result = append(result, c)
		} else if c := newCommand(code); c != nil {
			result = append(result, c)
		} else {
			break
		}
	}

	return CodeBlock{Name: name, Codes: result}
}

// Run execute CommandStep and return CodeResult
func (c Command) Run(s *ShellSession) CodeResult {
	rc, result := s.Run(strings.Join(c.Command, "\n") + "\n")
	o := CommandResult{Command: c, ActualOutput: strings.Split(result, "\n"), Rc: rc}
	return o
}

// Run execute TemplateCommand and return CodeResult
func (c *TemplateCommand) Run(s *ShellSession) CodeResult {
	_, result := s.Run(strings.Join(c.TemplateCommand, "\n") + "\n")
	cb := ParseCodeBlock("", strings.Split(result, "\n"))
	c.Codes = cb.Codes

	return c.Command.Run(s)
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
	return c.Command.StringLines()
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
