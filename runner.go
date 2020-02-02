package main

import (
	"log"
	"github.com/sergi/go-diff/diffmatchpatch"
)


func RunTestStep(s *Session, v *TestStep) {
    command := v.GetCommand()
    log.Println("Running", command)
    result := s.ExecuteCommand(command)
    expect := v.GetOutput()

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(result, expect, false)
	log.Println(dmp.DiffToDelta(diffs))
	log.Println(dmp.DiffPrettyText(diffs))
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
