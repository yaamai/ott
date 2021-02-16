package main

import (
	"bytes"
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

func getNearestHeading(fileBytes []byte, targetNode ast.Node) string {
	n := targetNode
	for limit := 0; limit < 3; limit++ {
		if head, ok := n.PreviousSibling().(*ast.Heading); ok {
			text := head.Lines().At(0)
			return string(text.Value(fileBytes))
		}
		n = n.PreviousSibling()
	}
	return ""
}

func walkCodeBlocks(source []byte, f func(n ast.Node, lines []string) []string) (ast.Node, []byte) {
	// prepare parser
	gm := newGoldmark()
	asts := gm.Parser().Parse(text.NewReader(source))

	// walk and modify asts
	ast.Walk(asts, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if (n.Kind() == ast.KindFencedCodeBlock || n.Kind() == ast.KindCodeBlock) && !entering {
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
