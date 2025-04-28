package main

import (
	"bytes"
	"github.com/Kunde21/markdownfmt/v3/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"log"
)

type CodeBlockTransformer struct{}

func (t *CodeBlockTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n := node.(type) {
		case *ast.FencedCodeBlock:
			log.Println(n.Info)
			log.Println(n.Text(reader.Source()))
			log.Println(reader.Source())
		}

		return ast.WalkContinue, nil
	})
}

func WalkMarkdown() {
	md := goldmark.New(
		goldmark.WithRenderer(markdown.NewRenderer()),
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.Prioritized(&CodeBlockTransformer{}, 0)),
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(),
	)
	source := []byte("# 111\n```\n# echo a\na\n```")
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		panic(err)
	}
	log.Println(buf.String())
}

func main() {
	WalkMarkdown()
}
