package main

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// RegexpRcMatch matches expressions for return code based checker
	RegexpRcMatch = regexp.MustCompile(`\(rc\s*((==|!=|<|>|<=|>=)([0-9]+))?\)$`)
)

// Checker is check CommandStepResult output
type Checker interface {
	IsMatch(result CommandStepResult) bool
}

// RcChecker is return-code based checker
type RcChecker struct {
	oper  string
	value int
}

// NewRcChecker creates RcChecker
func NewRcChecker(l string) *RcChecker {
	if rcMatch := RegexpRcMatch.FindStringSubmatch(l); rcMatch != nil {
		if rcMatch[2] == "" {
			return &RcChecker{oper: "==", value: 0}
		}

		rc, err := strconv.Atoi(rcMatch[3])
		if err != nil {
			return nil
		}
		return &RcChecker{oper: rcMatch[2], value: rc}
	}
	return nil
}

// IsMatch checks results has expected return code
func (m RcChecker) IsMatch(result CommandStepResult) bool {
	switch m.oper {
	case "==":
		return m.value == result.Rc
	case "!=":
		return m.value != result.Rc
	case "<=":
		return m.value <= result.Rc
	case ">=":
		return m.value >= result.Rc
	case "<":
		return m.value < result.Rc
	case ">":
		return m.value > result.Rc
	default:
		return false
	}
}

// HasChecker is return-code based checker
type HasChecker struct {
	ptn      string
	regexPtn *regexp.Regexp
}

// NewHasChecker creates HasChecker
func NewHasChecker(l string) *HasChecker {
	if !strings.HasSuffix(l, " (has)") {
		return nil
	}

	ptn := strings.TrimSuffix(l, " (has)")
	if strings.HasPrefix(ptn, "/") && strings.HasSuffix(ptn, "/") {
		t := strings.Split(ptn, "/")
		re, err := regexp.Compile(t[1])
		if err != nil {
			return nil
		}
		return &HasChecker{regexPtn: re}
	}

	return &HasChecker{ptn: ptn}
}

// IsMatch checks results has expected return code
func (m HasChecker) IsMatch(result CommandStepResult) bool {
	for _, l := range result.ActualOutput {
		if m.regexPtn != nil && m.regexPtn.MatchString(l) {
			return true
		}
		if m.regexPtn == nil && m.ptn == l {
			return true
		}
	}
	return false
}
