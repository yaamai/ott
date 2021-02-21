package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitByPrefixes(t *testing.T) {

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
	p := splitByPrefixes(data, []string{"<< ", "$ ", "# "}, []string{"<< ", "$ ", "# ", "> "})
	assert.Equal(t, 4, len(p))
	assert.Equal(t, map[string][]string{"<< ": {"aaa"}, "> ": {"bbb", "ccc"}}, p[0])
	assert.Equal(t, map[string][]string{"$ ": {"ddd"}, "": {"hhh"}}, p[1])
	assert.Equal(t, map[string][]string{"# ": {"eee"}, "": {""}}, p[2])
	assert.Equal(t, map[string][]string{"# ": {"fff"}, "> ": {"ggg"}, "": {"iii"}}, p[3])
}
