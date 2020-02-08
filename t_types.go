package main

type Line interface {
	Type() string
	Line() string
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

type EmptyLine struct {
	string
}

func (e *EmptyLine) Type() string {
	return "empty"
}
func (e *EmptyLine) Line() string {
	return e.string
}

type MetaCommentLine struct {
	string
}

func (c *MetaCommentLine) Type() string {
	return "meta-comment"
}
func (c *MetaCommentLine) Line() string {
	return c.string
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

type CommandLine struct {
	string
}

func (c *CommandLine) Type() string {
	return "command"
}
func (c *CommandLine) Line() string {
	return c.string
}

type OutputLine struct {
	string
}

func (c *OutputLine) Type() string {
	return "output"
}
func (c *OutputLine) Line() string {
	return c.string
}

type CommandContinueLine struct {
	string
}

func (c *CommandContinueLine) Type() string {
	return "output"
}
func (c *CommandContinueLine) Line() string {
	return c.string
}
