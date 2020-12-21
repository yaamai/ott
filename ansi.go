package main

import (
	"strings"
	"unicode/utf8"
)

const (
	AUTO_EXPAND_COL = 80
	AUTO_EXPAND_ROW = 24
)

type Cursor struct {
	x, y int
	lx   map[int]int
	ly   int
}

func NewCursor() *Cursor {
	return &Cursor{
		x: 0, y: 0,
		lx: map[int]int{},
		ly: 0,
	}
}

func (c *Cursor) Up(n int) {
	c.y -= n
	c.ly = c.y
}

func (c *Cursor) Down(n int) {
	c.y += n
	c.ly = c.y
}

func (c *Cursor) Left(n int) {
	if n == -1 {
		c.x = 0
	} else {
		c.x -= n
		c.lx[c.y] = c.x
	}
}

func (c *Cursor) Right(n int) {
	c.x += n
	c.lx[c.y] = c.x
}

func (c *Cursor) Clear(n int) {
	// FIXME: impl n=0,1
	c.lx[c.y] = 0
}

type Terminal struct {
	data   [][]rune
	c      *Cursor
	sx, sy int
}

func NewTerminal(sx, sy int) *Terminal {
	data := make([][]rune, sy)
	for i := range data {
		data[i] = make([]rune, sx)
	}

	return &Terminal{
		data: data,
		c:    NewCursor(),
		sx:   sx,
		sy:   sy,
	}
}

func RuneStringWithoutNull(ra []rune) string {
	// NOTE: string([]rune) outputs "\x00" runes.
	//       need to consinder write length (hold cursor pos?)
	for idx, r := range ra {
		if r == '\x00' {
			return string(ra[:idx])
		}
	}

	return ""
}

func StripEmptyLines(in []string) []string {
	for idx := len(in) - 1; idx >= 0; idx-- {
		if len(in[idx]) != 0 {
			return in[:idx+1]
		}
	}
	return []string{}
}

func (t Terminal) StringLines() []string {
	ret := []string{}

	for idx := range t.data[:t.c.ly+1] {
		ret = append(ret, string(t.data[idx][:t.c.lx[idx]]))
	}

	return ret
}

func (t Terminal) String() string {
	return strings.Join(t.StringLines(), "\n")
}

func GetCSISequence(buf []byte) int {
	for idx, c := range buf {
		if c >= 0x40 && c <= 0x7e {
			return idx + 1
		}
	}
	return 0
}

func (t *Terminal) alloc() {
	curRow := len(t.data)
	if t.c.y >= curRow {
		buf := make([][]rune, curRow+AUTO_EXPAND_ROW)
		copy(buf, t.data)
		t.data = buf
	}

	curCol := len(t.data[t.c.y])
	if t.c.x >= curCol {
		buf := make([]rune, curCol+AUTO_EXPAND_COL)
		copy(buf, t.data[t.c.y])
		t.data[t.c.y] = buf
	}
}

func (t *Terminal) Data(r rune) {
	t.alloc()
	t.data[t.c.y][t.c.x] = r
}

func (t *Terminal) Write(buf []byte) {
	for len(buf) > 0 {
		switch buf[0] {
		case '\x1b':
			switch buf[1] {
			case '\x5b':
				l := GetCSISequence(buf)
				n := buf[2] - 0x30
				typ := buf[3]
				switch typ {
				case '\x41':
					// Cursor Up
					t.c.Up(int(n))
				case '\x42':
					// Cursor Down
					t.c.Down(int(n))
				case '\x43':
					// Cursor Down
					t.c.Right(int(n))
				case '\x44':
					// Cursor Down
					t.c.Left(int(n))
				case '\x4b':
					// clear line
					t.c.Clear(int(n))
				}

				buf = buf[2+l:]
			}
		case '\x0d':
			t.c.Left(-1)
			buf = buf[1:]
		case '\x0a':
			t.c.Down(1)
			t.c.Left(-1)

			buf = buf[1:]
			t.alloc()
		default:
			r, size := utf8.DecodeRune(buf)
			t.Data(r)
			t.c.Right(1)

			buf = buf[size:]
		}
	}
}
