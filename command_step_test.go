package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	data := []string{
		"# fff",
		"> ggg",
		"iii",
	}
	c := newCommand(data)
	assert.Equal(t, Command{}, c)
}

func TestParseCodeBlock(t *testing.T) {
	data := []string{
		"<< aaa",
		"> bbb",
		"> ccc",
		"$ ddd",
		"hhh",
		"# eee",
		"",
		"# fff",
		"> ggg",
		"iii",
	}
	c := ParseCodeBlock("", data)
	assert.Equal(t, CodeBlock{
		Name: "",
		Codes: []Code{
			&TemplateCommand{TemplateCommand: []string{"aaa", "bbb", "ccc"}},
			&Command{Command: []string{"ddd"}, Output: []string{"hhh"}, Checker: []CommandChecker{}},
			&Command{Command: []string{"eee"}, Output: []string{""}, Checker: []CommandChecker{}},
			&Command{Command: []string{"fff", "ggg"}, Output: []string{"iii"}, Checker: []CommandChecker{}},
		},
	}, c)
}

func TestParseCodeBlock2(t *testing.T) {
	data := []string{
		"# ddd",
		"",
		"# eee",
		"",
		"# fff",
		"",
	}
	c := ParseCodeBlock("", data)
	assert.Equal(t, CodeBlock{
		Name: "",
		Codes: []Code{
			&Command{Command: []string{"ddd"}, Output: []string{""}, Checker: []CommandChecker{}},
			&Command{Command: []string{"eee"}, Output: []string{""}, Checker: []CommandChecker{}},
			&Command{Command: []string{"fff"}, Output: []string{""}, Checker: []CommandChecker{}},
		},
	}, c)
}
