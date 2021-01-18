package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Kunde21/markdownfmt/v2/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type CommandStep struct {
	Name    string
	Command []string
	Output  []string
}

type CommandStepResult struct {
	CommandStep
	ActualOutput []string
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

type Cli struct {
	sess  *ShellSession
	quiet bool
}

func NewCli(quiet bool, outputs string) (*Cli, error) {
	opts := []func(s *ShellSessionOption){}
	if !quiet {
		opts = append(opts, Mirror(os.Stderr))
	}
	sess, err := NewShellSession(opts...)
	if err != nil {
		return nil, err
	}

	return &Cli{
		sess:  sess,
		quiet: quiet,
	}, nil
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
	results := []CommandStepResult{}

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
		for _, s := range steps {
			c.onTestStepStart(name, s)
			r := s.Run(c.sess)
			c.onTestStepEnd(name, r)
			results = append(results, r)
		}
		return nil
	})

	c.onFileEnd(filename, fileBytes, modified, results)
	return results, nil
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
		quiet   = flag.Bool("q", false, "quiet")
		outputs = flag.String("o", "", "outputs")
	)
	flag.Parse()

	cli, err := NewCli(*quiet, *outputs)
	if err != nil {
		fmt.Println(err)
	}
	cli.RunFiles(flag.Args())
}
