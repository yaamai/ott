package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionInitialize(t *testing.T) {
	for idx := 0; idx < 100; idx += 1 {
		sess, err := NewSession("sh", "shell")
		if err != nil {
			t.Fatalf("err %s", err)
		}
		if string(sess.GetPrompt()) != "# " {
			t.Fatalf("(%d) prompt broken [%s]", idx, sess.GetPrompt())
		}
		sess.Cleanup()
	}
}

func TestExecuteCommand(t *testing.T) {
	sess, err := NewSession("sh", "shell")
	if err != nil {
		t.Fatalf("err %s", err)
	}

	output := sess.ExecuteCommand([]string{"echo a"})
	assert.Equal(t, []string{"a", ""}, output)

	output = sess.ExecuteCommand([]string{"echo a"})
	assert.Equal(t, []string{"a", ""}, output)

	// output = sess.ExecuteRawCommand("echo MARKER; echo a &&\\\n echo b; echo MARKER\n")
}

func TestExecuteCommandStability(t *testing.T) {
	sess, err := NewSession("sh", "shell")
	if err != nil {
		t.Fatalf("err %s", err)
	}

	for idx := 0; idx < 50; idx += 1 {
		output := sess.ExecuteCommand([]string{"echo a"})
		if output[0] != "a" {
			t.Fatalf("(%d) want =%s, got = %s", idx, "a", output)
		}
	}
}
