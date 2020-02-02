package main

import (
	"log"
)


func RunTestStep(s *Session, v *TestStep) {
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
