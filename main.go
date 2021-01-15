package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/Kunde21/markdownfmt/v2/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

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

type CommandStep struct {
	Command []string
	Output  []string
}

type CommandStepResult struct {
	CommandStep
	ActualOutput []string
}

func NewCommandSteps(lines []string) []CommandStep {
	steps := []CommandStep{}
	s := CommandStep{}

	for _, l := range lines {
		if strings.HasPrefix(l, "# ") {
			if len(s.Command) > 0 {
				steps = append(steps, s)
				s = CommandStep{}
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

func walkCodeBlocks(source []byte, f func(lines []string) []string) (ast.Node, []byte) {
	// prepare parser
	gm := newGoldmark()
	asts := gm.Parser().Parse(text.NewReader(source))

	// walk and modify asts
	ast.Walk(asts, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == ast.KindFencedCodeBlock && !entering {
			result := f(convertSegmentsToStringList(source, n.Lines()))
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

func run(sess *ShellSession, source []byte) ([]byte, []CommandStepResult) {
	results := []CommandStepResult{}
	_, modified := walkCodeBlocks(source, func(lines []string) []string {
		steps := NewCommandSteps(lines)
		for _, s := range steps {
			results = append(results, s.Run(sess))
		}
		log.Println(steps, results)
		return nil
	})

	return modified, results
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

// TODO: mirror command output with prefix
// TODO: parse command steps name (using textblock ast before codeblock)
// TODO: write actual output md
// TODO: rewrite cli output
func main() {

	flag.Parse()

	sess, err := NewShellSession(Mirror(os.Stderr, []byte("  # "), []byte("  ")))
	if err != nil {
		log.Println(err)
	}
	results := map[string][]CommandStepResult{}
	for _, arg := range flag.Args() {
		bytes, err := ioutil.ReadFile(arg)
		if err != nil {
			log.Println(err)
		}

		_, r := run(sess, bytes)
		log.Println(formatCommandStepResults(arg, r))
		results[arg] = r
	}

	// source := []byte("---\nTitle: 100\n---\n\n# test `# a:100`\n```\n# aaaa\n# bbbb\ncccc\n```\n# aaa\n[]: # aaa")
	// fmt.Print(string(source))
	// run(source)
}
