package main

import (
	"bytes"
	"io"
	"strconv"
)

// Reader is flexible byte array reader
type Reader struct {
	base       io.Reader
	buf        *bytes.Buffer
	rpos, wpos int
}

// NewReader creates Reader with underlay io.Reader
func NewReader(size int, r io.Reader) *Reader {
	b := &bytes.Buffer{}
	b.Grow(size)
	return &Reader{
		base: r,
		buf:  b,
	}
}

// ReadWithFunc reads bytes array with specified condition
func (r *Reader) ReadWithFunc(f func([]byte, int64) (int64, []byte)) []byte {
	for {
		l, _ := r.buf.ReadFrom(r.base)

		l, output := f(r.buf.Bytes(), l)
		if output != nil {
			r.buf.Reset()
			return output
		}
	}
}

// ReadToPattern reads bytes array to specified patterns founds
func (r *Reader) ReadToPattern(pattern []byte) []byte {
	return r.ReadWithFunc(func(buf []byte, l int64) (int64, []byte) {
		pos := bytes.Index(buf, pattern)
		if pos == -1 {
			return 0, nil
		}

		return int64(pos), buf[:pos]
	})
}

// MultiPatternParser is pattern based bytes array parser
type MultiPatternParser struct {
	startPattern, endPattern     [][]byte
	dataCb                       func(data []byte)
	startPatternCb, endPatternCb func(data []byte, pos [][2]int64)
}

// Parse implements Reader's Parse function
func (p *MultiPatternParser) Parse(buf []byte, l int64) (int64, []byte) {
	pos := indexMultiple(buf, p.startPattern, p.endPattern)
	startPos := int64(-1)
	if len(pos) > 0 {
		startPos = pos[0][len(pos[0])-1][1]
	}
	endPos := int64(-1)
	if len(pos) > 1 {
		endPos = pos[1][0][0]
	}

	p.callDataCallback(buf, l, startPos, endPos)
	p.callPatternCallback(buf, pos)

	if startPos == -1 || endPos == -1 {
		return 0, nil
	}
	return endPos - startPos, buf[startPos:endPos]
}

// TODO: if read partial pattern, callback may incorrect
func (p *MultiPatternParser) callDataCallback(buf []byte, l int64, startPos int64, endPos int64) {
	if p.dataCb == nil {
		return
	}

	cbStartPos := int64(len(buf)) - l
	if startPos != -1 && cbStartPos < startPos {
		cbStartPos = startPos
	}
	if startPos != -1 && endPos == -1 {
		if p.dataCb != nil && int64(len(buf)) > startPos {
			p.dataCb(buf[cbStartPos:])
		}
	}

	if endPos != -1 && endPos > cbStartPos {
		p.dataCb(buf[cbStartPos:endPos])
	}
}

func (p *MultiPatternParser) callPatternCallback(buf []byte, pos [][][2]int64) {
	if len(pos) > 0 && p.startPatternCb != nil {
		p.startPatternCb(buf, pos[0])
	}
	if len(pos) > 1 && p.endPatternCb != nil {
		p.endPatternCb(buf, pos[1])
	}
}

// ShellParser is shell-output adjusted MultiPatternParser
type ShellParser struct {
	MultiPatternParser
	rc     int
	mirror io.Writer
}

// NewShellParser creates ShellParser
func NewShellParser(marker [][]byte, mirror io.Writer) *ShellParser {
	p := ShellParser{MultiPatternParser: MultiPatternParser{startPattern: marker, endPattern: marker}, mirror: mirror}
	p.dataCb = p.mirrorData
	p.endPatternCb = p.parseReturnCode
	return &p
}

func (p *ShellParser) mirrorData(data []byte) {
	if p.mirror != nil {
		p.mirror.Write(data)
	}
}

func (p *ShellParser) parseReturnCode(data []byte, pos [][2]int64) {
	i, err := strconv.Atoi(string(data[pos[0][1]:pos[1][0]]))
	if err != nil {
		p.rc = -1
	} else {
		p.rc = i
	}
}
