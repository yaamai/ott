package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Kunde21/markdownfmt/v2/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type CommandStep struct {
	Name    string   `json:"name"`
	Command []string `json:"command"`
	Output  []string `json:"output"`
}

type CommandStepResult struct {
	CommandStep
	ActualOutput []string `json:"actual"`
}

func NewCommandSteps(name string, lines []string) []CommandStep {
	steps := []CommandStep{}
	s := CommandStep{}

	for _, l := range lines {
		if strings.HasPrefix(l, "# ") {
			if len(s.Command) > 0 {
				steps = append(steps, s)
				s = CommandStep{Name: name}
			}
			s.Command = append(s.Command, strings.TrimPrefix(l, "# "))
		} else if strings.HasPrefix(l, "> ") {
			s.Command = append(s.Command, strings.TrimPrefix(l, "> "))
		} else {
			s.Output = append(s.Output, l)
		}
	}
	if len(s.Command) > 0 {
		steps = append(steps, s)
	}

	return steps
}

func (c CommandStep) Run(s *ShellSession) CommandStepResult {
	result := s.Run(strings.Join(c.Command, "\n") + "\n")
	o := CommandStepResult{CommandStep: c, ActualOutput: strings.Split(result, "\n")}
	return o
}

func (c CommandStepResult) IsOutputsExpected() bool {
	if len(c.Output) != len(c.ActualOutput) {
		return false
	}

	for idx := range c.Output {
		if c.Output[idx] != c.ActualOutput[idx] {
			return false
		}
	}

	return true
}

func (c CommandStepResult) StringLines() []string {
	result := []string{}
	prompt := "#"
	for _, cmd := range c.Command {
		result = append(result, fmt.Sprintf("%s %s\n", prompt, cmd))
		prompt = ">"
	}
	for _, out := range c.ActualOutput {
		result = append(result, fmt.Sprintf("%s\n", out))
	}

	return result
}

func newGoldmark() goldmark.Markdown {
	mr := markdown.NewRenderer()

	extensions := []goldmark.Extender{
		extension.GFM,
	}
	parserOptions := []parser.Option{
		parser.WithAttribute(), // We need this to enable # headers {#custom-ids}.
	}
	gm := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(parserOptions...),
	)
	gm.SetRenderer(mr)
	return gm
}

func convertSegmentsToStringList(source []byte, s *text.Segments) []string {
	lines := []string{}
	for idx := 0; idx < s.Len(); idx++ {
		text := s.At(idx)
		lines = append(lines, strings.TrimSuffix(string(text.Value(source)), "\n"))
	}
	return lines
}

func createNewSegment(source *[]byte, s string) text.Segment {
	start := len(*source)
	stop := start + len(s)
	*source = append(*source, []byte(s)...)

	return text.NewSegment(start, stop)
}

func createNewSegments(source *[]byte, l []string) *text.Segments {
	s := text.NewSegments()
	for _, e := range l {
		s.Append(createNewSegment(source, e))
	}

	return s
}

func walkCodeBlocks(source []byte, f func(n ast.Node, lines []string) []string) (ast.Node, []byte) {
	// prepare parser
	gm := newGoldmark()
	asts := gm.Parser().Parse(text.NewReader(source))

	// walk and modify asts
	ast.Walk(asts, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == ast.KindFencedCodeBlock && !entering {
			result := f(n, convertSegmentsToStringList(source, n.Lines()))
			if result != nil {
				n.SetLines(createNewSegments(&source, result))
			}
		}
		return ast.WalkContinue, nil
	})

	buf := bytes.Buffer{}
	gm.Renderer().Render(&buf, source, asts)
	return asts, buf.Bytes()
}

func countCommandStepResults(results []CommandStepResult) (string, int, int) {
	success := 0
	fail := 0
	for _, step := range results {
		if step.IsOutputsExpected() {
			success += 1
		} else {
			fail += 1
		}
	}

	s := "OK"
	if success != fail {
		s = "FAIL"
	}
	return s, success, fail
}

func convertCommandStepResults(results []CommandStepResult) []string {
	result := []string{}
	for _, r := range results {
		result = append(result, r.StringLines()...)
	}
	return result
}

func formatCommandStepResults(name string, results []CommandStepResult) string {
	var s string = name + ": "
	for _, step := range results {
		if step.IsOutputsExpected() {
			s += "."
		} else {
			s += "!"
		}
	}

	return s
}

// TODO: write actual output md
// TODO: rewrite cli output
/*
default:
example.t.md:
  => test A
  # echo a
  a
  => test B
  # echo b
=> FAIL (1/2)

-q:
example.t.md: !.... FAIL (1/5)

-o json
  example.t.md.json
-o md:onerr
  example.t.err.md
-o md:always
  example.t.md

md-err
md-all
json
*/
// TODO: ansi

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
		fmt.Printf("%s", b.String())
	}
}

func (c *Cli) onFileStart(filename string) {
	fmt.Printf("%s: ", filename)
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
