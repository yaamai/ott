package main

import (
	"testing"
)

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
