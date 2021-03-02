package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkAnyCodes(t *testing.T) {
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
	actual := []interface{}{}
	walkAnyCodes(data, func(c interface{}) {
		actual = append(actual, c)
	})

	expected := []interface{}{
		&TemplateCommand{TemplateCommand: []string{"aaa", "bbb", "ccc"}},
		&Command{Command: []string{"ddd"}, Output: []string{"hhh"}, Checker: []CommandChecker{}},
		&Command{Command: []string{"eee"}, Output: []string{""}, Checker: []CommandChecker{}},
		&Command{Command: []string{"fff", "ggg"}, Output: []string{"iii"}, Checker: []CommandChecker{}},
	}
	assert.Equal(t, expected, actual)
}
