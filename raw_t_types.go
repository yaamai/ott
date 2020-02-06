package main

type Line interface {
	Type() string
	Line() string
	Equal(l Line) bool
}

type CommentLine struct {
	string
}

func (c *CommentLine) Type() string {
	return "comment"
}
func (c *CommentLine) Line() string {
	return c.string
}
func (c *CommentLine) Equal(l Line) bool {
	return c.Line() == l.Line()
}

type MetaCommentLine struct {
	string
	parent *MetaCommentLine
}

func (c *MetaCommentLine) Type() string {
	return "meta-comment"
}
func (c *MetaCommentLine) Line() string {
	return c.string
}
func (c *MetaCommentLine) Equal(l Line) bool {
	return c.Line() == l.Line()
}

type TestCaseLine struct {
	string
}

func (c *TestCaseLine) Type() string {
	return "test-case"
}
func (c *TestCaseLine) Line() string {
	return c.string
}
func (c *TestCaseLine) Equal(l Line) bool {
	return c.Line() == l.Line()
}

type TestCaseCommentLine struct {
	string
	parent *TestCaseLine
}

func (c *TestCaseCommentLine) Type() string {
	return "test-case-comment"
}
func (c *TestCaseCommentLine) Line() string {
	return c.string
}
func (c *TestCaseCommentLine) Equal(l Line) bool {
	return c.Line() == l.Line()
}

type CommandLine struct {
	string
	parent *TestCaseLine
}

func (c *CommandLine) Type() string {
	return "command"
}
func (c *CommandLine) Line() string {
	return c.string
}
func (c *CommandLine) Equal(l Line) bool {
	return c.Line() == l.Line()
}

type OutputLine struct {
	string
	parent *CommandLine
}

func (c *OutputLine) Type() string {
	return "output"
}
func (c *OutputLine) Line() string {
	return c.string
}
func (c *OutputLine) Equal(l Line) bool {
	return c.Line() == l.Line()
}

type CommandContinueLine struct {
	string
	parent *CommandLine
}

func (c *CommandContinueLine) Type() string {
	return "output"
}
func (c *CommandContinueLine) Line() string {
	return c.string
}
func (c *CommandContinueLine) Equal(l Line) bool {
	return c.Line() == l.Line()
}
