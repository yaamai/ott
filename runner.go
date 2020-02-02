package main

import (
	"log"
	"fmt"
    "github.com/pmezard/go-difflib/difflib"
)


func RunTestStep(s *Session, v *TestStep) {
    command := v.GetCommand()
    log.Println("Running", command)
    result := s.ExecuteCommand(command)
    expect := v.GetOutput()


    log.Println("R", result, expect)
    diff := difflib.UnifiedDiff{
        A:        difflib.SplitLines(expect),
        B:        difflib.SplitLines(result),
        FromFile: "Expected",
        ToFile:   "Output",
        Context:  3,
    }
    text, _ := difflib.GetUnifiedDiffString(diff)
    fmt.Printf(text)
}

func RunTestCase(s *Session, v *TestCase) {
    log.Println(v.Name)
    log.Println(v.Metadata)
    log.Println(v.TestSteps)
    for _, step := range(v.TestSteps) {
        RunTestStep(s, step)
    }
}

func Run(s *Session, t *TFile) {
	for _, line := range(t.Lines) {

        switch v := line.(type) {
            case *Comment:
			    log.Println(v)
            case *TestCase:
                RunTestCase(s, v)
        }
    }
}
