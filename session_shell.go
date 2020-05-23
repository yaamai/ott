package main

import (
	"bytes"
	"os"
)

var (
	SHELL_START_MARKER       = `###OTT-START###`
	SHELL_START_MARKER_BYTES = []byte(SHELL_START_MARKER)
	SHELL_START_MARKER_CMD   = []byte(`echo -n "` + SHELL_START_MARKER + `"; `)
	SHELL_END_MARKER         = `###OTT-END###`
	SHELL_END_MARKER_CMD     = []byte(`; echo -n "` + SHELL_END_MARKER + `"`)
	SHELL_END_MARKER_BYTES   = []byte(SHELL_END_MARKER)
	LF                       = []byte("\n")
)

type ShellSession struct {
	prompt []byte
}

func (s *ShellSession) GuessPrompt(buffer *LockedBuffer, ptmx *os.File) []byte {
	prompt := guessPrompt(buffer, ptmx)
	s.prompt = prompt
	return prompt
}

func (s *ShellSession) GetPrompt() []byte {
	return s.prompt
}

func (s *ShellSession) GetCmdline(cmdStrs []string) []byte {
	cmds := getBytesArray(cmdStrs)
	cmdline := make([]byte, 0)

	cmdline = append(cmdline, SHELL_START_MARKER_CMD...)
	cmdline = append(cmdline, bytes.Join(cmds, LF)...)
	cmdline = append(cmdline, SHELL_END_MARKER_CMD...)
	cmdline = append(cmdline, LF...) // generate marker output

	return cmdline
}

func (s *ShellSession) GetStartMarker(cmdStrs []string) []byte {
	return SHELL_START_MARKER_BYTES
}

func (s *ShellSession) GetEndMarker(_ []string) []byte {
	return SHELL_END_MARKER_BYTES
}

func (s *ShellSession) NormalizeOutput(output []byte) []string {
	return getStringArray(bytes.Split(output, []byte("\r\n")))
}
