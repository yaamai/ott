package main

import (
	"bytes"
	"os"
)

type PythonSession struct {
	prompt []byte
}

func (s *PythonSession) GuessPrompt(buffer *LockedBuffer, ptmx *os.File) []byte {
	prompt := guessPrompt(buffer, ptmx)
	prompt = bytes.TrimSuffix(bytes.TrimPrefix(prompt, []byte("\r\n")), []byte("\r\n"))
	s.prompt = prompt
	return prompt
}

func (s *PythonSession) GetPrompt() []byte {
	return s.prompt
}

func (s *PythonSession) GetCmdline(cmdStrs []string) []byte {
	cmds := getBytesArray(cmdStrs)
	cmdline := bytes.Join(cmds, []byte("\n"))
	cmdline = append(cmdline, []byte("\n\n")...)

	return cmdline
}

func (s *PythonSession) GetStartMarker(cmdStrs []string) []byte {
	cmds := getBytesArray(cmdStrs)
	expectStartMarker := bytes.Join(cmds, []byte("\r\n... "))
	expectStartMarker = append(expectStartMarker, []byte("\r\n")...)
	return expectStartMarker
}

func (s *PythonSession) GetEndMarker(_ []string) []byte {
	return s.prompt
}

func (s *PythonSession) NormalizeOutput(output []byte) []string {
	return getStringArray(bytes.Split(output, []byte("\r\n")))
}
