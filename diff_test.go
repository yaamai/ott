package main

import (
	"fmt"
	"testing"
)

func TestGetMarkedCommand(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"b", "b", "c"}
	//a = []string{"k", "i", "t", "t", "e", "n"}
	//b = []string{"s", "i", "t", "t", "i", "n", "g"}
	//a = []string{"a", "b", "c", "a", "b", "b", "a"}
	//b = []string{"c", "b", "a", "b", "a", "c"}
	a = []string{"a", "b", "c", "a", "b", "b", "a"}
	b = []string{"c", "b", "a", "b", "a", "c"}
	d := calc_diff(a, b)
	fmt.Println(d)
}
