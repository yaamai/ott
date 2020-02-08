package main

import (
	"github.com/pmezard/go-difflib/difflib"
	"go.uber.org/zap"
)

type Runner struct {
	session *Session
}

func NewRunner() (*Runner, error) {
	sess, err := NewSession()
	if err != nil {
		return nil, err
	}

	return &Runner{
		session: sess,
	}, nil
}

func (r *Runner) Run(testFile *TestFile) {
	zap.S().Info("Running test-file: ", testFile.Name)
	for testCaseIdx, testCase := range testFile.Tests {
		zap.S().Info("Running test-case: ", testCase.Name)
		for testStepIdx, testStep := range testCase.Steps {
			zap.S().Debug("Running test-step: ", testStep.Command)
			zap.S().Debug("                 : ", testStep.ExpectedOutput)
			actualOutput := r.session.ExecuteCommand(testStep.Command)
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(testStep.ExpectedOutput),
				B:        difflib.SplitLines(actualOutput),
				FromFile: "Expected",
				ToFile:   "Output",
				Context:  3,
			}
			text, _ := difflib.GetUnifiedDiffString(diff)

			testFile.Tests[testCaseIdx].Steps[testStepIdx].ActualOutput = actualOutput
			testFile.Tests[testCaseIdx].Steps[testStepIdx].Diff = text
			zap.S().Debug(text)
		}
	}
}
