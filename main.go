package main

import (
	"flag"
	"fmt"
	"strings"
)

// StringList implements array flags
type StringList []string

func (i *StringList) String() string {
	return strings.Join(*i, ",")
}

// Set implements flags.Set() interface
func (i *StringList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var (
		quiet   bool
		outputs StringList = []string{
			"json:ok:logs/{{.TS}}/{{.BaseName}}.{{.Format}}",
			"md:always:logs/{{.TS}}/{{.BaseName}}.{{.Format}}",
		}
	)
	flag.BoolVar(&quiet, "q", false, "quiet")
	flag.Var(&outputs, "o", "outputs")
	flag.Parse()

	cli, err := NewCli(quiet, outputs)
	if err != nil {
		fmt.Println(err)
	}
	cli.RunFiles(flag.Args())
}
