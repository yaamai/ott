package main

import (
	"testing"
    "fmt"
)

/*
func TestGetMarkedCommand(t *testing.T) {
	tests := []struct {
		c    string
		want string
	}{
		{"", "\n ### OTT-OTT ###\necho -n ### OTT-OTT ###\n"},
		{"date", "\ndate ### OTT-OTT ###\necho -n ### OTT-OTT ###\n"},
		{"date &&\\ date", "\ndate &&\\ date ### OTT-OTT ###\necho -n ### OTT-OTT ###\n"},
	}
	for _, tt := range tests {
		if got := getMarkedCommand(tt.c); string(got) != tt.want {
			t.Fatalf("want = %s, got = %s", tt.want, string(got))
		}
	}
}
*/


func TestSessionInitialize(t *testing.T) {
    for idx := 0; idx < 100; idx += 1 {
        sess, err := NewSession()
        if err != nil {
            t.Fatalf("err %s", err)
        }
        if string(sess.Prompt) != "sh-5.0$ " {
            t.Fatalf("prompt broken [%s]", sess.Prompt)
        }
        fmt.Println("idx", idx)
        sess.Cleanup()
    }
}

/*
func TestExecuteCommand(t *testing.T) {
    sess, err := NewSession()
    if err != nil {
        t.Fatalf("err %s", err)
    }

    output := sess.ExecuteCommand("echo a")
    fmt.Println([]byte(output))
    if output != "a\n" {
        t.Fatalf("want =%s, got = %s", "a", output)
    }
}

func TestExecuteCommandStability(t *testing.T) {
    sess, err := NewSession()
    if err != nil {
        t.Fatalf("err %s", err)
    }

    for idx := 0; idx < 100; idx += 1 {
        output := sess.ExecuteCommand("echo a")
        fmt.Println("idx: ", idx, "output:", []byte(output))
        if output != "a\n" {
            t.Fatalf("want =%s, got = %s", "a", output)
        }
    }
}
*/
