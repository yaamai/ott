package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"golang.org/x/term"
)

type ShellSession struct {
	ptmx   *os.File
	reader *Reader
}

func NewShellSession(opts) (*ShellSession, error) {
	c := exec.Command("bash")
	c.Env = append(c.Env, "PS1=###")
	c.Env = append(c.Env, "HISTFILE=/dev/null")
	c.Args = append(c.Args, []string{"--norc", "--noprofile"}...)

	winsize := pty.Winsize{Cols: 80, Rows: 24}
	ptmx, err := pty.StartWithSize(c, &winsize)
	if err != nil {
		return nil, err
	}
	term.MakeRaw(int(ptmx.Fd()))

	reader := NewReader(65536, ptmx)

	return &ShellSession{ptmx: ptmx, reader: reader}, nil
}

type Reader struct {
	base       io.Reader
	buf        []byte
	rpos, wpos int
}

func NewReader(size int, r io.Reader) *Reader {
	return &Reader{
		base: r,
		buf:  make([]byte, size),
	}
}

func (r *Reader) ReadWithFunc(f func([]byte) (int, []byte)) []byte {
	for {
		l, _ := r.base.Read(r.buf[r.wpos:])
		r.wpos += l

		l, output := f(r.buf[r.rpos:r.wpos])
		if output != nil {
			r.rpos += l
			return output
		}
	}
}

func (r *Reader) ReadToPattern(pattern []byte) []byte {
	return r.ReadWithFunc(func(buf []byte) (int, []byte) {
		pos := bytes.Index(buf, pattern)
		if pos == -1 {
			return 0, nil
		}

		return pos, buf[:pos]
	})
}

func (r *Reader) ReadBetweenPattern(startPattern, endPattern []byte) []byte {
	return r.ReadWithFunc(func(buf []byte) (int, []byte) {
		startPos := bytes.Index(buf, startPattern)
		if startPos == -1 {
			return 0, nil
		}
		startPos += len(startPattern)
		endPos := bytes.Index(buf[startPos:], endPattern)
		if endPos == -1 {
			return 0, nil
		}
		endPos += startPos

		return endPos - startPos, buf[startPos:endPos]
	})
}

func (s *ShellSession) Run(cmd string) string {
	pattern := []byte("###")

	s.ptmx.Write([]byte("eval \"echo -n ===; exit $?\"\n"))
	s.reader.ReadToPattern([]byte("==="))

	s.ptmx.Write([]byte(cmd))
	return strings.TrimSuffix(string(s.reader.ReadBetweenPattern(pattern, pattern)), "\n")
}
