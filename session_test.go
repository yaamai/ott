package main

import (
	"testing"
)


func TestGetMarkedCommand(t *testing.T) {
	tests := []struct {
		c    string
		want string
	}{
		{"", "; PS1=###OTT-OTT###\n"},
		{"date", "date; PS1=###OTT-OTT###\n"},
		{"date &&\\ date", "date &&\\ date; PS1=###OTT-OTT###\n"},
	}
	for _, tt := range tests {
		if got := getMarkedCommand(tt.c); string(got) != tt.want {
			t.Fatalf("want = %s, got = %s", tt.want, string(got))
		}
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
    if output != "a\n" {
        t.Fatalf("want =%s, got = %s", "a", output)
    }

    output = sess.ExecuteCommand("echo a")
    if output != "a\n" {
        t.Fatalf("want =%s, got = %s", "a", output)
    }
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
