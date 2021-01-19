package main

import (
	"bytes"
	"fmt"
	"strings"

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
