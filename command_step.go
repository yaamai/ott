package main

import (
	"fmt"
	"strings"
)

// CommandStep represents a command-line
type CommandStep struct {
	Name    string    `json:"name"`
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

// NewCommandSteps parses command-step string arrays to CommandStep
func NewCommandSteps(name string, lines []string) []CommandStep {
	steps := []CommandStep{}
	s := CommandStep{}

	for _, l := range lines {
		if strings.HasPrefix(l, "# ") {
			if len(s.Command) > 0 {
				steps = append(steps, s)
				s = CommandStep{Name: name}
			}
			s.Command = append(s.Command, strings.TrimPrefix(l, "# "))
		} else if strings.HasPrefix(l, "> ") {
			s.Command = append(s.Command, strings.TrimPrefix(l, "> "))
		} else {
			if m := NewRcChecker(l); m != nil {
				s.Checker = append(s.Checker, m)
			} else if m := NewHasChecker(l); m != nil {
				s.Checker = append(s.Checker, m)
			} else {
				s.Output = append(s.Output, l)
			}
		}
	}
	if len(s.Command) > 0 {
		steps = append(steps, s)
	}

	return steps
}

// Run execute CommandStep and return CommandStepResult
func (c CommandStep) Run(s *ShellSession) CommandStepResult {
	rc, result := s.Run(strings.Join(c.Command, "\n") + "\n")
	o := CommandStepResult{CommandStep: c, ActualOutput: strings.Split(result, "\n"), Rc: rc}
	return o
}

// IsOutputsExpected checks CommandStepResult is expected outputs or not
func (c CommandStepResult) IsOutputsExpected() bool {
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
		if step.IsOutputsExpected() {
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