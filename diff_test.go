package main

import (
	"fmt"
	"testing"
	"strings"
	"regexp"
)

func TestCalcDiff(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"b", "b", "c"}
	a = []string{"k", "i", "t", "t", "e", "n"}
	b = []string{"s", "i", "t", "t", "i", "n", "g"}
	a = []string{"a", "b"}
	b = []string{"a", "b"}
	a = []string{"aa", "b1"}
	b = []string{"a\\d (re)", "b\\D (re)"}
	//a = []string{"a", "b", "c", "a", "b", "b", "a"}
	//b = []string{"c", "b", "a", "b", "a", "c"}
	//a = []string{"a", "b", "c", "a", "b", "b", "a"}
	//b = []string{"c", "b", "a", "b", "a", "c"}
	d := calcDiff(b, a, func(expect, actual string) bool {
		match := false
		if strings.HasSuffix(expect, " (re)") {
			match, _ = regexp.MatchString(strings.TrimSuffix(expect, " (re)"), actual)
		} else {
			match = expect == actual
		}
		return match
	})
	fmt.Println(strings.Join(d, "\n"))
}
