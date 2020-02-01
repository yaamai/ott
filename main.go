package main

import (
	"log"
)

// TODO:
// Session
// ShellParser
// Runner
// TestFile
// TestCase
// TestStep

func main() {
	s, err := NewSession()
	if err != nil {
		log.Fatalln(err)
	}

	r := s.ExecuteCommand("date &&\\\ndate #")
	log.Println(r)
}
