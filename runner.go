package main

import (
	"fmt"
	"github.com/pmezard/go-difflib/difflib"
	"go.uber.org/zap"
)

func RunTestStep(s *Session, v *TestStep) {
	command := v.GetCommand()
	zap.S().Debug("Running", command)
	result := s.ExecuteCommand(command)
	expect := v.GetOutput()

	zap.S().Debug("R", result, expect)
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(expect),
		B:        difflib.SplitLines(result),
		FromFile: "Expected",
		ToFile:   "Output",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	fmt.Printf(text)
}

func RunTestCase(s *Session, v *TestCase) {
	zap.S().Debug(v.Name)
	zap.S().Debug(v.Metadata)
	zap.S().Debug(v.TestSteps)
	for _, step := range v.TestSteps {
		RunTestStep(s, step)
	}
}

func Run(s *Session, t *TFile) {
	for _, line := range t.Lines {

		switch v := line.(type) {
		case *Comment:
			zap.S().Debug(v)
		case *TestCase:
			RunTestCase(s, v)
		}
	}
}
