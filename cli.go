package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
)

// Cli represents CLI
type Cli struct {
	sess    *ShellSession
	quiet   bool
	outputs []string
}

// NewCli create Cli instance
func NewCli(quiet bool, outputs []string) (*Cli, error) {
	opts := []func(s *ShellSessionOption){}
	if !quiet {
		opts = append(opts, Mirror(os.Stderr))
	}
	sess, err := NewShellSession(opts...)
	if err != nil {
		return nil, err
	}

	return &Cli{
		sess:    sess,
		quiet:   quiet,
		outputs: outputs,
	}, nil
}

// TemplateContext represents tests output filename template context datas
type TemplateContext struct {
	FileName string
	BaseName string
	Format   string
	Timing   string
	Time     time.Time
	TS       string
}

func parseOutputFlag(filename, out string) (string, TemplateContext) {
	list := strings.Split(out, ":")
	format := list[0]

	timing := "always"
	if len(list) > 1 {
		timing = list[1]
	}

	templateStr := "{{.Filename}}{{.Format}}"
	if len(list) > 2 {
		templateStr = list[2]
	}

	ctx := TemplateContext{
		filename,
		strings.TrimSuffix(filename, filepath.Ext(filename)),
		format,
		timing,
		time.Now(),
		time.Now().Format("20060102_150405"),
	}
	return templateStr, ctx
}

func (c *Cli) outputFiles(origFilename string, input, output []byte, results []CodeResult) {
	for _, out := range c.outputs {
		var b strings.Builder
		templateStr, ctx := parseOutputFlag(origFilename, out)
		templ, _ := template.New("filename").Parse(templateStr)
		templ.Execute(&b, ctx)
		outFilename := b.String()

		_, _, fail := countCommandStepResults(results)
		if (ctx.Timing == "always") || (ctx.Timing == "err" && fail != 0) || (ctx.Timing == "ok" && fail == 0) {
			dir := filepath.Dir(outFilename)
			os.MkdirAll(dir, 0755)

			if ctx.Format == "md" {
				ioutil.WriteFile(outFilename, output, 0644)
			}
			if ctx.Format == "json" {
				bytes, _ := json.Marshal(results)
				ioutil.WriteFile(outFilename, bytes, 0644)
			}
		}
	}
}

func (c *Cli) onFileStart(filename string) {
	color.New(color.FgCyan, color.Bold).Printf("== %s ==\n", filename)
	if !c.quiet {
		fmt.Printf("\n")
	}
}

func (c *Cli) onFileEnd(filename string, input, output []byte, results []CodeResult) {
	if !c.quiet {
		fmt.Printf("=>")
	}
	s, success, _ := countCommandStepResults(results)
	fmt.Printf(" %s (%d/%d)\n", s, success, len(results))

	c.outputFiles(filename, input, output, results)
}

func (c *Cli) onCodeBlockStart(cbName string) {}
func (c *Cli) onCodeBlockEnd(cbName string)   {}

func (c *Cli) onCodeStart(cbName string, step Code) {
	if !c.quiet {
		color.New(color.FgCyan).Printf("%s:\n", cbName)
		fmt.Print(strings.Join(step.StringLines(), ""))
	}
}

func (c *Cli) onCodeEnd(cbName string, step CodeResult) {
	if c.quiet {
		if step.Check() {
			fmt.Print(".")
		} else {
			fmt.Print("!")
		}
	} else {
		if step.Check() {
			color.New(color.FgGreen).Printf("=> OK\n")
		} else {
			color.New(color.FgRed).Printf("=> FAIL\n")
		}
		fmt.Print("\n")
	}
}

// RunFile runs file-based tests
func (c *Cli) RunFile(filename string) ([]CodeResult, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return RunMarkdown(filename, fileBytes, c.sess, c), nil
}

// RunFiles runs multiple file-based tests
func (c *Cli) RunFiles(filenames []string) (map[string][]CodeResult, error) {
	results := map[string][]CodeResult{}

	for _, filename := range filenames {
		c.RunFile(filename)
	}

	return results, nil
}
