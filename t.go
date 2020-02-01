package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"regexp"
)

type Lineable interface {
	Type() string
	Lines() []string
}

type TFile struct {
	Lines []Lineable
}

var (
	CommentLineRe          = regexp.MustCompile(`^\s*#.*$`)
	MetaCommentLineRe      = regexp.MustCompile(`^\s*# meta.*$`)
	MetaDataLineRe         = regexp.MustCompile(`^\s*#\s+(.*?):\s*(.*?)$`)
	TestCaseLineRe         = regexp.MustCompile(`^.*:\s*$`)
	TestStepLineRe         = regexp.MustCompile(`^  \$ .*$`)
	TestStepContinueLineRe = regexp.MustCompile(`^  > .*$`)
	TestStepOutputLineRe   = regexp.MustCompile(`^  [^>$].*$`)
)

func ParseTFile(stream io.Reader) (TFile, error) {
	t := TFile{}
	scanner := bufio.NewScanner(stream)
	var currentTestCase *TestCase
	var currentTestStep *TestStep
	var currentTestCaseMeta *TestMeta
	// TODO: introduce Context, function-tables to cleanup code
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)

		// parse test case meta
		if MetaCommentLineRe.MatchString(line) {
			currentTestCaseMeta = &TestMeta{String: line, Meta: map[string]string{}}
			continue
		}
		if currentTestCaseMeta != nil && CommentLineRe.MatchString(line) {
			match := MetaDataLineRe.FindStringSubmatch(line)
			key := match[1]
			value := match[2]
			currentTestCaseMeta.Meta[key] = value
			continue
		}

		// parse normal comment
		if currentTestCaseMeta == nil && CommentLineRe.MatchString(line) {
			c := Comment{line}
			t.Lines = append(t.Lines, &c)
			continue
		}

		// detect test case start
		if TestCaseLineRe.MatchString(line) {
			testCase := TestCase{Name: line}
			if currentTestCaseMeta != nil {
				testCase.Metadata = *currentTestCaseMeta
				currentTestCaseMeta = nil
			}
			t.Lines = append(t.Lines, &testCase)
			currentTestCase = &testCase
			continue
		}
		if currentTestCase != nil {
			if TestStepLineRe.MatchString(line) {
				testStep := TestStep{Commands: []string{line}}
				currentTestCase.TestSteps = append(currentTestCase.TestSteps, &testStep)
				currentTestStep = &testStep
				continue
			}
			if currentTestStep != nil {
				if TestStepContinueLineRe.MatchString(line) {
					currentTestStep.Commands = append(currentTestStep.Commands, line)
					continue
				}
				if TestStepOutputLineRe.MatchString(line) {
					currentTestStep.Output = append(currentTestStep.Output, line)
				}
			}
		}
	}
	return t, nil
}

func (t *TFile) UnmarshalJSON(b []byte) error {
	// parse json array only
	m := []json.RawMessage{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	// detect "type" value
	typeStruct := struct{ Type string }{}
	for _, line := range m {
		err := json.Unmarshal(line, &typeStruct)
		if err != nil {
			return err
		}

		var dst interface{}
		switch typeStruct.Type {
		case "comment":
			dst = new(Comment)
		case "testcase":
			dst = new(TestCase)
		}

		// unmarshal to dedicated type
		err = json.Unmarshal(line, dst)
		if err != nil {
			return err
		}
		t.Lines = append(t.Lines, dst.(Lineable))
	}
	log.Println(m)

	return nil
}

type Comment struct {
	String string `json:"string"`
}

func (c *Comment) Type() string {
	return "comment"
}
func (c *Comment) Lines() []string {
	return []string{c.String}
}

type TestCase struct {
	Metadata  TestMeta    `json:"metadata"`
	Name      string      `json:"name"`
	TestSteps []*TestStep `json:"steps"`
}

func (t *TestCase) Type() string {
	return "testcase"
}
func (t *TestCase) Lines() []string {
	return []string{t.Name}
}

type TestMeta struct {
	String string            `json:"string"`
	Meta   map[string]string `json:"meta"`
}

type TestStep struct {
	Commands []string `json:"commands"`
	Output   []string `json:"outputs"`
}
