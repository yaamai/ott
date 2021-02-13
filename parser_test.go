package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRcChecker(t *testing.T) {
	m := NewRcChecker("(rc==0)")
	assert.NotNil(t, m)
	assert.True(t, m.IsMatch())

	m = NewRcChecker("(rc)")
	assert.NotNil(t, m)
	assert.True(t, m.IsMatch())
}
