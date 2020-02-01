package main

import (
	"log"
	"bufio"
	"io"
	"regexp"
)


type Lineable interface {
	Lines() []string
	Equal(l Lineable) bool
}

type TFile struct {
	Lines []Lineable
}

var (
	CommentLineRe = regexp.MustCompile(`^\s*#.*$`)
)

func ParseTFile(stream io.Reader) (TFile, error) {
	t := TFile{}
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)
		if CommentLineRe.MatchString(line) {
			t.Lines = append(t.Lines, Comment(line))
		}
	}
	return t, nil
}

type Comment string
func (c Comment) Lines() []string {
	return []string{string(c)}
}

func (c Comment) Equal(l Lineable) bool {
	return c.Lines() == l.Lines()
}

type TestCase struct {
	Metadata TestMeta
	Name string
	TestSteps []TestStep
}

type TestMeta struct {}

type TestStep struct {
	Commands []Command
}

type Command string
