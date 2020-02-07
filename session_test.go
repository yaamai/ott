package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMarkedCommand(t *testing.T) {
	tests := []struct {
		c    string
		want string
	}{
		{"", "echo -n \"###OTT-START###\"; ; echo -n \"###OTT-END###\"\n"},
		{"date", "echo -n \"###OTT-START###\"; date; echo -n \"###OTT-END###\"\n"},
		{"date &&\\ date", "echo -n \"###OTT-START###\"; date &&\\ date; echo -n \"###OTT-END###\"\n"},
	}
	for _, tt := range tests {
		got := getMarkedCommand(tt.c)
		assert.Equal(t, tt.want, string(got))
	}
}

func TestSessionInitialize(t *testing.T) {
	for idx := 0; idx < 100; idx += 1 {
		sess, err := NewSession()
		if err != nil {
			t.Fatalf("err %s", err)
		}
		if string(sess.GetPrompt()) != "sh-5.0$ " {
			t.Fatalf("(%d) prompt broken [%s]", idx, sess.GetPrompt())
		}
		sess.Cleanup()
	}
}

func TestExecuteCommand(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatalf("err %s", err)
	}

	output := sess.ExecuteCommand("echo a")
	assert.Equal(t, "a\n", output)

	output = sess.ExecuteCommand("echo a")
	assert.Equal(t, "a\n", output)

	// output = sess.ExecuteRawCommand("echo MARKER; echo a &&\\\n echo b; echo MARKER\n")
}

func TestExecuteCommandStability(t *testing.T) {
	sess, err := NewSession()
	if err != nil {
		t.Fatalf("err %s", err)
	}

	for idx := 0; idx < 50; idx += 1 {
		output := sess.ExecuteCommand("echo a")
		if output != "a\n" {
			t.Fatalf("(%d) want =%s, got = %s", idx, "a", output)
		}
	}
}
