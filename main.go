package main

import (
	//        "io"
	"log"
	//        "os/signal"
	//        "syscall"
	//        "golang.org/x/crypto/ssh/terminal"
	//	"strings"
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

	s.ExecuteCommand("date &&\\\ndate #")
}
