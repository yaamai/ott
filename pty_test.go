package main

import (
	// "log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexMultiple(t *testing.T) {
	tests := []struct {
		desc string
		buf  []byte
		in   [][]byte
		out  [][][2]int
	}{
		{"empty", []byte{}, [][]byte{}, [][][2]int{}},
		{"normal", []byte("AAABAAA"), [][]byte{[]byte("B")}, [][][2]int{{{3, 4}}}},
		{"two-pattern", []byte("AAABABAA"), [][]byte{[]byte("B"), []byte("B")}, [][][2]int{{{3, 4}, {5, 6}}}},
		{"long-pattern", []byte("AAABBAAA"), [][]byte{[]byte("AAA"), []byte("AAA")}, [][][2]int{{{0, 3}, {5, 8}}}},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := indexMultiple(tt.buf, tt.in)
			assert.Equal(t, tt.out, p)
		})
	}
}

func TestMultiPatternParser(t *testing.T) {
	buf := make([]byte, 0, 32)
	cbBuf := [][]byte{}
	cbFn := func(data []byte) { cbBuf = append(cbBuf, data) }
	ptn := [][]byte{[]byte("###"), []byte("###")}
	dataIn := [][]byte{[]byte("###"), []byte("100###"), []byte("D"), []byte("ATA"), []byte("###0###")}

	p := MultiPatternParser{ptn, ptn, cbFn, nil, nil}
	for idx, d := range dataIn {
		buf = append(buf, d...)
		_, data := p.Parse(buf, len(d))

		if idx == len(dataIn)-1 {
			expectCb := [][]byte{[]byte("D"), []byte("ATA")}
			assert.Equal(t, []byte("DATA"), data)
			assert.Equal(t, expectCb, cbBuf)
		} else {
			assert.Nil(t, data)

		}
	}
}

func TestExecuteCommandStability(t *testing.T) {
	sess, err := NewShellSession()
	assert.Nil(t, err)
	assert.NotNil(t, sess)

	for idx := 0; idx < 1; idx++ {
		_, output := sess.Run("echo a\n")
		// log.Println(idx, output)
		assert.Equal(t, "a", output)
	}
}

func TestFailureCommand(t *testing.T) {
	sess, err := NewShellSession()
	assert.Nil(t, err)
	assert.NotNil(t, sess)

	_, output := sess.Run(";\n")
	assert.Equal(t, "sh: syntax error near unexpected token `;'", output)
}
