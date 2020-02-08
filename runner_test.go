package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInsert(t *testing.T) {
	tests := []struct {
		src    []*TestCase
		expect []*TestCase
		d      *TestCase
        idx    int
		err    error
	}{
		{[]*TestCase{}, []*TestCase{&TestCase{Generated: true}}, &TestCase{}, 0, nil},
		{[]*TestCase{&TestCase{Name: "a"}}, []*TestCase{&TestCase{Generated: true}, &TestCase{Name: "a"}}, &TestCase{}, 0, nil},
		{[]*TestCase{&TestCase{Name: "a"}}, []*TestCase{&TestCase{Name: "a"}, &TestCase{Generated: true}}, &TestCase{}, 1, nil},
		{[]*TestCase{&TestCase{Name: "a"}, &TestCase{Name: "b"}}, []*TestCase{&TestCase{Name: "a"}, &TestCase{Generated: true}, &TestCase{Name: "b"}}, &TestCase{Generated: true}, 1, nil},
	}
	for _, tt := range tests {
        got := insert(tt.src, tt.idx, tt.d)

		assert.Equal(t, tt.expect, got)
	}
}
