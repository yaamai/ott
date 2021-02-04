package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/yuin/goldmark/ast"
	"github.com/fatih/color"
)

// TODO: ansi
// TODO: matcher(re)
// TODO: matcher(rc)
// TODO: matcher(ignore)
// TODO: matcher(has)

type StringList []string

func (i *StringList) String() string {
	return strings.Join(*i, ",")
}

func (i *StringList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type Cli struct {
	sess    *ShellSession
	quiet   bool
	outputs []string
}

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

func (c *Cli) outputFiles(origFilename string, input, output []byte, results []CommandStepResult) {
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
	// fmt.Printf("%s: ", filename)
	color.Cyan("%s: ", filename)
	if !c.quiet {
		fmt.Printf("\n")
	}
}

func (c *Cli) onFileEnd(filename string, input, output []byte, results []CommandStepResult) {
	if !c.quiet {
		fmt.Printf("=>")
	}
	s, success, _ := countCommandStepResults(results)
	fmt.Printf(" %s (%d/%d)\n", s, success, len(results))

	c.outputFiles(filename, input, output, results)
}

func (c *Cli) onTestStepStart(stepname string, step CommandStep) {
	if !c.quiet {
		fmt.Printf("  %s:\n", stepname)
		prompt := "#"
		for _, cmd := range step.Command {
			fmt.Printf("  %s %s\n", prompt, cmd)
			prompt = ">"
		}
	}
}

func (c *Cli) onTestStepEnd(stepname string, step CommandStepResult) {
	if c.quiet {
		if step.IsOutputsExpected() {
			fmt.Print(".")
		} else {
			fmt.Print("!")
		}
	}
}

func (c *Cli) RunFile(filename string) ([]CommandStepResult, error) {
	fileResults := []CommandStepResult{}

	c.onFileStart(filename)
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	_, modified := walkCodeBlocks(fileBytes, func(n ast.Node, lines []string) []string {
		var name string
		if prev, ok := n.(*ast.FencedCodeBlock).PreviousSibling().(*ast.Heading); ok {
			text := prev.Lines().At(0)
			name = string(text.Value(fileBytes))
		}

		steps := NewCommandSteps(name, lines)
		stepsResults := []CommandStepResult{}
		for _, s := range steps {
			c.onTestStepStart(name, s)
			r := s.Run(c.sess)
			c.onTestStepEnd(name, r)
			stepsResults = append(stepsResults, r)
		}
		fileResults = append(fileResults, stepsResults...)
		return convertCommandStepResults(stepsResults)
	})

	c.onFileEnd(filename, fileBytes, modified, fileResults)
	return fileResults, nil
}

func (c *Cli) RunFiles(filenames []string) (map[string][]CommandStepResult, error) {
	results := map[string][]CommandStepResult{}

	for _, filename := range filenames {
		c.RunFile(filename)
	}

	return results, nil
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
