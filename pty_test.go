package main

import (
	// "log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommandStability(t *testing.T) {
	sess, err := NewShellSession()
	assert.Nil(t, err)
	assert.NotNil(t, sess)

	for idx := 0; idx < 100; idx += 1 {
		output := sess.Run("echo a\n")
		// log.Println(idx, output)
		assert.Equal(t, "a", output)
	}
}

func TestFailureCommand(t *testing.T) {
	sess, err := NewShellSession()
	assert.Nil(t, err)
	assert.NotNil(t, sess)

	output := sess.Run(";\n")
	assert.Equal(t, "bash: syntax error near unexpected token `;'", output)
}
