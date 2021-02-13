package main

import (
	"strings"
	"unicode/utf8"
)

const (
	// AutoExpandCols is default expand size
	AutoExpandCols = 80
	// AutoExpandRows is default expand size
	AutoExpandRows = 24
)

// Cursor represents terminal cursor
type Cursor struct {
	x, y int
	lx   map[int]int
	ly   int
}

// NewCursor creates Cursor
func NewCursor() *Cursor {
	return &Cursor{
		x: 0, y: 0,
		lx: map[int]int{},
		ly: 0,
	}
}

// Up do a cursor up
func (c *Cursor) Up(n int) {
	c.y -= n
	c.ly = c.y
}

// Down do a cursor down
func (c *Cursor) Down(n int) {
	c.y += n
	c.ly = c.y
}

// Left do a cursor left
func (c *Cursor) Left(n int) {
	if n == -1 {
		c.x = 0
	} else {
		c.x -= n
		c.lx[c.y] = c.x
	}
}

// Right do a cursor right
func (c *Cursor) Right(n int) {
	c.x += n
	c.lx[c.y] = c.x
}

// Clear do a cursor position clear (to line headings)
func (c *Cursor) Clear(n int) {
	// FIXME: impl n=0,1
	c.lx[c.y] = 0
}

// Terminal represents vt100 based terminal (partially)
type Terminal struct {
	data   [][]rune
	c      *Cursor
	sx, sy int
}

// NewTerminal creates Terminal
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

// StringLines converts terminal data to string array
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

// GetCSISequence calculates CSI sequences length
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
		buf := make([][]rune, curRow+AutoExpandRows)
		copy(buf, t.data)
		t.data = buf
	}

	curCol := len(t.data[t.c.y])
	if t.c.x >= curCol {
		buf := make([]rune, curCol+AutoExpandCols)
		copy(buf, t.data[t.c.y])
		t.data[t.c.y] = buf
	}
}

// Data puts rune to terminal
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
					t.c.Up(int(n))
				case '\x42':
					t.c.Down(int(n))
				case '\x43':
					t.c.Right(int(n))
				case '\x44':
					t.c.Left(int(n))
				case '\x4b':
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
