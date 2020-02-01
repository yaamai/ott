package main

import (
	"log"
	"bufio"
	"io"
	"encoding/json"
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
	CommentLineRe = regexp.MustCompile(`^\s*#.*$`)
	TestCaseLineRe = regexp.MustCompile(`^.*:\s*$`)
	TestStepLineRe = regexp.MustCompile(`^  \$ .*$`)
	TestStepContinueLineRe = regexp.MustCompile(`^  > .*$`)
)

func ParseTFile(stream io.Reader) (TFile, error) {
	t := TFile{}
	scanner := bufio.NewScanner(stream)
	var currentTestCase *TestCase
	var currentTestStep *TestStep
	// TODO: introduce Context, function-tables to cleanup code
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)
		if CommentLineRe.MatchString(line) {
			c := Comment{line}
			t.Lines = append(t.Lines, &c)
			continue
		}
		if TestCaseLineRe.MatchString(line) {
			testCase := TestCase{Name: line}
			t.Lines = append(t.Lines, &testCase)
			currentTestCase = &testCase
			continue
		}
		if currentTestCase != nil {
			if TestStepLineRe.MatchString(line) {
				testStep := TestStep{Commands: []Command{Command(line)}}
				currentTestCase.TestSteps = append(currentTestCase.TestSteps, &testStep)
				currentTestStep = &testStep
				continue
			}
			if currentTestStep != nil {
				if TestStepContinueLineRe.MatchString(line) {
					currentTestStep.Commands = append(currentTestStep.Commands, Command(line))
					continue
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
	typeStruct := struct{Type string}{}
	for _, line := range(m) {
		err := json.Unmarshal(line, &typeStruct)
		if err != nil {
			return err
		}

		var dst interface{}
		switch (typeStruct.Type) {
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
	Metadata TestMeta `json:"metadata"`
	Name string `json:"name"`
	TestSteps []*TestStep `json:"steps"`
}
func (t *TestCase) Type() string {
	return "testcase"
}
func (t *TestCase) Lines() []string {
	return []string{t.Name}
}

type TestMeta struct {}

type TestStep struct {
	Commands []Command `json:"commands"`
}

type Command string
