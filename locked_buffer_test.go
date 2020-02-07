package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLockedBuffer_ReadBetweenPattern(t *testing.T) {
	startPattern := []byte("START")
	endPattern := []byte("END")
	tests := []struct {
		buf  string
		want string
	}{
		{"STARTaaaa\nbbbb\nEND", "aaaa\nbbbb\n"},
	}
	for _, tt := range tests {
		b := LockedBuffer{}
		b.Write([]byte(tt.buf))
		want, err := b.ReadBetweenPattern(startPattern, endPattern)
		assert.Nil(t, err)
		assert.Equal(t, []byte(tt.want), want)
	}
}
