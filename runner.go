package main

import (
	"go.uber.org/zap"
	"strings"
	"regexp"
)

type Runner struct {
	session *Session
}

func NewRunner(cmd, mode string) (*Runner, error) {
	sess, err := NewSession(cmd, mode)
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


func diff(expect, actual []string) []string {
	diff := []string{}

	for idx, _ := range expect {
		zap.S().Debug(expect[idx], actual[idx], expect[idx] == actual[idx])
		match := false
		if strings.HasSuffix(expect[idx], " (re)") {
			match, _ = regexp.MatchString(strings.TrimSuffix(expect[idx], " (re)"), actual[idx])
		} else {
			match = expect[idx] == actual[idx]
		}
		zap.S().Debug(match)
		if match {
			diff = append(diff, actual[idx])
		} else {
			diff = append(diff, "- " + expect[idx])
			diff = append(diff, "+ " + actual[idx])
		}
	}

	return diff
}

func (r *Runner) runTestCase(testCase *TestCase) {
	for testStepIdx, testStep := range testCase.Steps {
		zap.S().Debug("Running test-step: ", testStep.Command)
		zap.S().Debug("                 : ", testStep.ExpectedOutput)

		actualOutput := r.session.ExecuteCommand(testStep.Command)
		// remove new-line at end. t-file can't represent new-line at end.
		if len(actualOutput) > 0 {
			if actualOutput[len(actualOutput)-1] == "" {
				actualOutput = actualOutput[:len(actualOutput)-1]
			}
		}
		testCase.Steps[testStepIdx].Diff = diff(testStep.ExpectedOutput, actualOutput)
		testCase.Steps[testStepIdx].ActualOutput = actualOutput[0]
	}
}

func getPrefixedTestCase(testFile *TestFile, prefixes ...string) [][]*TestCase {
	result := [][]*TestCase{}

	for _, prefix := range prefixes {

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
	for _, c := range src {
		copiedTestCase := *c
		copiedTestCase.Generated = true
		(*dest) = append((*dest), &copiedTestCase)
	}
}

func getMarkedPrefixedTestCase(testFile *TestFile, prefixes ...string) [][]*TestCase {
	testCasesList := getPrefixedTestCase(testFile, prefixes...)

	// mark prefixed test-cases as NoInject
	for _, testCases := range testCasesList {
		for _, testCase := range testCases {
			testCase.NoInject = true
		}
	}
	return testCasesList
}

func getHookTestCases(testFileList []*TestFile) ([][]*TestCase, [][]*TestCase) {
	// TODO: consider test-cases as container/list
	setupTestCases := [][]*TestCase{[]*TestCase{}, []*TestCase{}, []*TestCase{}}
	teardownTestCases := [][]*TestCase{[]*TestCase{}, []*TestCase{}, []*TestCase{}}

	for _, testFile := range testFileList {
		c := getMarkedPrefixedTestCase(testFile, "setup-per-run", "setup-per-file", "setup-per-case")
		for idx, _ := range c {
			setupTestCases[idx] = append(setupTestCases[idx], c[idx]...)
		}

		c = getMarkedPrefixedTestCase(testFile, "teardown-per-run", "teardown-per-file", "teardown-per-case")
		for idx, _ := range c {
			teardownTestCases[idx] = append(teardownTestCases[idx], c[idx]...)
		}
	}

	return setupTestCases, teardownTestCases
}

func injectTestCase(testFileList []*TestFile) []*TestFile {
	setupTestCases, teardownTestCases := getHookTestCases(testFileList)
	zap.S().Debug("Injecting setup/teardown test-cases: ", setupTestCases, teardownTestCases)

	injectedTestFiles := []*TestFile{}

	// inject per-run testcase as new TestFile
	setupRunTestFile := TestFile{Name: "generated", Generated: true}
	appendGeneratedTestCase(&setupRunTestFile.Tests, setupTestCases[0])
	injectedTestFiles = append(injectedTestFiles, &setupRunTestFile)

	for _, testFile := range testFileList {
		injectedTestCases := []*TestCase{}

		// inject per-file test-cases
		appendGeneratedTestCase(&injectedTestCases, setupTestCases[1])

		// insert per-case
		for _, targetTestCase := range testFile.Tests {
			if targetTestCase.Generated || targetTestCase.NoInject {
				continue
			}
			appendGeneratedTestCase(&injectedTestCases, setupTestCases[2])
			injectedTestCases = append(injectedTestCases, targetTestCase)
			appendGeneratedTestCase(&injectedTestCases, teardownTestCases[2])
		}

		appendGeneratedTestCase(&injectedTestCases, teardownTestCases[1])

		testFile.Tests = injectedTestCases
		injectedTestFiles = append(injectedTestFiles, testFile)
	}

	teardownRunTestFile := TestFile{Name: "generated", Generated: true}
	appendGeneratedTestCase(&teardownRunTestFile.Tests, teardownTestCases[0])
	injectedTestFiles = append(injectedTestFiles, &teardownRunTestFile)

	return injectedTestFiles
}

func (r *Runner) run(testFile *TestFile) {

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

func (r *Runner) RunMultiple(testFileList []*TestFile) []*TestFile {
	injectedTestFiles := injectTestCase(testFileList)

	for _, testFile := range injectedTestFiles {
		r.run(testFile)
	}

	return injectedTestFiles
}
