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

func appendGeneratedTestCase(dest *[]*TestCase, src []*TestCase) {
    for _, c := range(src) {
        copiedTestCase := *c
        copiedTestCase.Generated = true
        (*dest) = append((*dest), &copiedTestCase)
    }
}

func getMarkedPrefixedTestCase(testFile *TestFile, prefixes... string) [][]*TestCase {
    testCasesList := getPrefixedTestCase(testFile, prefixes...)

    // mark prefixed test-cases as NoInject
    for _, testCases := range(testCasesList) {
        for _, testCase := range(testCases) {
            testCase.NoInject = true
        }
    }
    return testCasesList
}

func injectTestCase(testFile *TestFile) {
    // TODO: consider test-cases as container/list
    setupTestCases := getMarkedPrefixedTestCase(testFile, "setup-per-run", "setup-per-file", "setup-per-case")
    teardownTestCases := getMarkedPrefixedTestCase(testFile, "teardown-per-run", "teardown-per-file", "teardown-per-case")
    zap.S().Debug("Injecting setup/teardown test-cases: ", setupTestCases, teardownTestCases)


    result := []*TestCase{}

    // insert per-run,per-file testcase
    // TODO: support multiple file
    appendGeneratedTestCase(&result, setupTestCases[0])
    appendGeneratedTestCase(&result, setupTestCases[1])

    // insert per-csae
    for _, targetTestCase := range(testFile.Tests) {
        if targetTestCase.Generated || targetTestCase.NoInject {
            continue
        }
        appendGeneratedTestCase(&result, setupTestCases[2])
        result = append(result, targetTestCase)
        appendGeneratedTestCase(&result, teardownTestCases[2])
    }

    appendGeneratedTestCase(&result, teardownTestCases[1])
    appendGeneratedTestCase(&result, teardownTestCases[0])

    (*testFile).Tests = result
}

func (r *Runner) Run(testFile *TestFile) {

    injectTestCase(testFile)

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
