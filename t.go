package main

import (
	"log"
	"bufio"
	"io"
	"regexp"
)


type Lineable interface {
	Lines() []string
}

type TFile struct {
	Lines []Lineable
}

var (
	CommentLineRe = regexp.MustCompile(`^\s*#.*$`)
	TestCaseLineRe = regexp.MustCompile(`^.*:\s*$`)
	TestStepLineRe = regexp.MustCompile(`^  [$>] .*$`)
)

func ParseTFile(stream io.Reader) (TFile, error) {
	t := TFile{}
	scanner := bufio.NewScanner(stream)
	var currentTestCase *TestCase
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)
		if CommentLineRe.MatchString(line) {
			c := Comment{line}
			t.Lines = append(t.Lines, &c)
		}
		if TestCaseLineRe.MatchString(line) {
			testCase := TestCase{Name: line}
			t.Lines = append(t.Lines, &testCase)
			currentTestCase = &testCase
		}
		if currentTestCase != nil {
			if TestStepLineRe.MatchString(line) {
				currentTestCase.TestSteps = append(currentTestCase.TestSteps, TestStep{Commands: []Command{Command(line)}})
			}
		}
	}
	return t, nil
}

type Comment struct {
	Line string
}
func (c *Comment) Lines() []string {
	return []string{c.Line}
}

type TestCase struct {
	Metadata TestMeta
	Name string
	TestSteps []TestStep
}
func (t *TestCase) Lines() []string {
	return []string{t.Name}
}

type TestMeta struct {}

type TestStep struct {
	Commands []Command
}

type Command string
