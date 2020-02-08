package main

import (
	"github.com/pmezard/go-difflib/difflib"
	"go.uber.org/zap"
    "strings"
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


func getTestCaseMap(testFile *TestFile) map[string]*TestCase {
    result := map[string]*TestCase{}
	for testCaseIdx, testCase := range testFile.Tests {
        result[testCase.Name] = testFile.Tests[testCaseIdx]
    }

    return result
}

func (r *Runner) runTestCase(testCase *TestCase) {
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

		testCase.Steps[testStepIdx].ActualOutput = actualOutput
		testCase.Steps[testStepIdx].Diff = text
		zap.S().Debug(text)
	}
}

func getPrefixedTestCase(testFile *TestFile, prefixes... string) [][]*TestCase {
    result := [][]*TestCase{}

    for _, prefix := range(prefixes) {

        cases := make([]*TestCase, 0)
	    for testCaseIdx, testCase := range testFile.Tests {
            if strings.HasPrefix(testCase.Name, prefix) {
                cases = append(cases, testFile.Tests[testCaseIdx])
            }
        }

        result = append(result, cases)
    }

    return result
}

func insert(testCaseSlice []*TestCase, idx int, testCase *TestCase) []*TestCase {
    // copy test-case
    testCaseCopy := *testCase
    testCaseCopy.Generated = true

    testCaseSlice = append(testCaseSlice, &TestCase{})
    copy(testCaseSlice[idx+1:], testCaseSlice[idx:])
    testCaseSlice[idx] = &testCaseCopy

    return testCaseSlice
}

func insertSetupTestCase(testFile *TestFile) {
    // TODO: consider test-cases as container/list
    prefixedTestCases := getPrefixedTestCase(testFile, "setup-per-run", "setup-per-file", "setup-per-case")
    zap.S().Debug("Injecting setup test-cases", prefixedTestCases)

    // mark prefixed test-cases as NoInject
    for _, testCases := range(prefixedTestCases) {
        for _, testCase := range(testCases) {
            testCase.NoInject = true
        }
    }

    result := []*TestCase{}

    // insert per-run testcase
    for _, c := range(prefixedTestCases[0]) {
        copiedTestCase := *c
        copiedTestCase.Generated = true
        result = append(result, &copiedTestCase)
    }

    // insert per-file testcase
    // TODO: support multiple file
    for _, c := range(prefixedTestCases[1]) {
        copiedTestCase := *c
        copiedTestCase.Generated = true
        result = append(result, &copiedTestCase)
    }

    // insert per-csae
    for _, targetTestCase := range(testFile.Tests) {
        if targetTestCase.Generated || targetTestCase.NoInject {
            continue
        }

        for _, c := range(prefixedTestCases[2]) {
            copiedTestCase := *c
            copiedTestCase.Generated = true
            result = append(result, &copiedTestCase)
        }

        result = append(result, targetTestCase)
    }

    (*testFile).Tests = result
}

func (r *Runner) Run(testFile *TestFile) {

    insertSetupTestCase(testFile)

	zap.S().Info("Running test-file: ", testFile.Name)
	for testCaseIdx, testCase := range testFile.Tests {
		zap.S().Info("Running test-case: ", testCase.Name)
        if strings.HasPrefix(testCase.Name, "setup-") && !testCase.Generated {
		    zap.S().Debug("skipping setup testcase: ", testCase.Name)
            continue
        }
        r.runTestCase(testFile.Tests[testCaseIdx])
	}
}
