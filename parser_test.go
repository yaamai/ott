package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRcChecker(t *testing.T) {
	m := NewRcChecker("(rc==0)")
	r := CommandStepResult{Rc: 0}
	assert.NotNil(t, m)
	assert.True(t, m.IsMatch(r))

	m = NewRcChecker("(rc)")
	r = CommandStepResult{Rc: 0}
	assert.NotNil(t, m)
	assert.True(t, m.IsMatch(r))

	m = NewRcChecker("(rc==1)")
	r = CommandStepResult{Rc: 0}
	assert.NotNil(t, m)
	assert.False(t, m.IsMatch(r))
}
