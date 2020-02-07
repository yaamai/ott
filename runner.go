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
	for _, testCase := range testFile.Tests {
		zap.S().Info("Running test-case: ", testCase.Name)
		for _, testStep := range testCase.Steps {
			zap.S().Debug("Running test-step: ", testStep.Command)
			zap.S().Debug("                 : ", testStep.Output)
			actualOutput := r.session.ExecuteCommand(testStep.Command)
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(testStep.Output),
				B:        difflib.SplitLines(actualOutput),
				FromFile: "Expected",
				ToFile:   "Output",
				Context:  3,
			}
			text, _ := difflib.GetUnifiedDiffString(diff)
			zap.S().Info(text)
		}
	}
}
