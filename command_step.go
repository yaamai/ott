package main

import (
	"fmt"
	"strings"
)

// CodeBlock represents a CodeBlock in markdown
type CodeBlock struct {
	Name  string `json:"name"`
	Codes []ShellRunnable
}

// ShellRunnable represents runnable in shell
type ShellRunnable interface {
	Run(s *ShellSession) ShellCheckable
}

// ShellCheckable represents checkable with shell
type ShellCheckable interface {
	Check() bool
}

// TemplateCommandStep represents a meta command-line
type TemplateCommandStep struct {
	CommandStep
	TemplateCommand []string `json:"template_command"`
}

// CommandStep represents a command-line
type CommandStep struct {
	Command []string  `json:"command"`
	Output  []string  `json:"output"`
	Checker []Checker `json:"checker"`
}

// CommandStepResult represents executed CommandStep results
type CommandStepResult struct {
	CommandStep
	ActualOutput []string `json:"actual"`
	Rc           int      `json:"rc"`
}

func ParseCodeBlock(name string, lines []string) CodeBlock {
	result := []ShellRunnable{}

	t := splitByPrefixes(lines, []string{"<< ", "$ ", "# "}, []string{"<< ", "$ ", "# ", "> "})
	for _, m := range t {
		if _, ok := m["<< "]; ok {
			result = append(result, TemplateCommandStep{TemplateCommand: append(m["<< "], m["> "]...)})
		} else if _, ok := m["# "]; ok {
			result = append(result, CommandStep{Command: append(m["# "], m["> "]...), Output: m[""]})
		} else if _, ok2 := m["$ "]; ok2 {
			result = append(result, CommandStep{Command: append(m["$ "], m["> "]...), Output: m[""]})
		}
	}

	return CodeBlock{Name: name, Codes: result}
}

// NewCommandStep parses command-step string arrays to CommandStep
func NewCommandStep(name string, lines []string) CommandStep {
	t := splitLines(lines, "# ", "> ")
	s := CommandStep{Name: name, Command: append(t[0], t[1]...)}
	for _, l := range t[2] {
		if m := NewRcChecker(l); m != nil {
			s.Checker = append(s.Checker, m)
		} else if m := NewHasChecker(l); m != nil {
			s.Checker = append(s.Checker, m)
		} else {
			s.Output = append(s.Output, l)
		}
	}

	return s
}

// Run execute CommandStep and return CommandStepResult
func (c CommandStep) Run(s *ShellSession) ShellCheckable {
	rc, result := s.Run(strings.Join(c.Command, "\n") + "\n")
	o := CommandStepResult{CommandStep: c, ActualOutput: strings.Split(result, "\n"), Rc: rc}
	return o
}

// Check checks CommandStepResult is expected outputs or not
func (c CommandStepResult) Check() bool {
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
func (c CommandStepResult) StringLines() []string {
	result := []string{}
	prompt := "#"
	for _, cmd := range c.Command {
		result = append(result, fmt.Sprintf("%s %s\n", prompt, cmd))
		prompt = ">"
	}
	for _, out := range c.ActualOutput {
		result = append(result, fmt.Sprintf("%s\n", out))
	}

	return result
}

func countCommandStepResults(results []CommandStepResult) (string, int, int) {
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

func convertCommandStepResults(results []CommandStepResult) []string {
	result := []string{}
	for _, r := range results {
		result = append(result, r.StringLines()...)
	}
	return result
}
