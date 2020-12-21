package main

import (
	"log"
	"strings"
	"unicode/utf8"
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
	log.Println(t.c.lx)

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
		default:
			r, size := utf8.DecodeRune(buf)
			t.data[t.c.y][t.c.x] = r
			t.c.Right(1)

			buf = buf[size:]
		}
	}
}

/*
Using default tag: latest
latest: Pulling from library/alpine
801bfaa63ef2: Pull complete
Digest: sha256:3c7497bf0c7af93428242d6176e8f7905f2201d8fc5861f45be7a346b5f23436
Status: Downloaded newer image for alpine:latest
docker.io/library/alpine:latest
*/

/*
import (
	"encoding/hex"
	"fmt"
	"github.com/creack/pty"
	"strings"
	// "golang.org/x/term"
	"log"
	"os/exec"
	"time"
	"unicode/utf8"
)


func GetCSISequence(buf []byte) int {
	if !(len(buf) >= 3 && buf[0] == 0x1b && buf[1] == 0x5b) {
		return 0
	}

	for idx, c := range buf[2:] {
		if c >= 0x40 && c <= 0x7e {
			return 2 + idx + 1
		}
	}
	return 0
}

func main() {
	data, err := hex.DecodeString(strings.Join([]string{
		"5573696e672064656661756c74207461673a206c61746573740a",
		"6c61746573743a2050756c6c696e672066726f6d206c6962726172792f616c70696e650a",
		"0a",
		"1b5b31411b5b324b0d3830316266616136336566323a2050756c6c696e67206673206c61796572200d",
	}, ""))
	log.Println(data, err)

	term := [24][80]rune{}
	cx, cy := 0, 0

	for len(data) > 0 {
		seq := GetCSISequence(data)
		if seq != 0 {
			fmt.Printf("  CSI: %x %d\n", data[:seq], seq)

			n := data[2] - 0x30
			t := data[3]
			if t == 0x41 {
				// Cursor Up
				cy -= int(n)
			} else if t == 0x42 {
				// Cursor Down
				cy += int(n)
			} else if t == 0x4b {
				cx = 0
			} else {
				log.Printf("ignored %x", t)
			}

			data = data[seq:]
			continue
		}

		r, size := utf8.DecodeRune(data)
		// fmt.Printf("%c", r)
		// log.Println(cx, cy)

		if cx >= 80 || cy >= 24 {
			log.Println("breaked")
			break
		}
		term[cy][cx] = r
		cx += 1
		// log.Printf("%x\n", data[0])
		if data[0] == 0x0a {
			cx = 0
			cy += 1
		}

		data = data[size:]
	}

	log.Println("========")
	for idx := 0; idx < 24; idx++ {
		log.Println(string(term[idx][:]))
	}

	// for len(data) > 0 {
	// 	r, size := utf8.DecodeRune(data)
	// 	fmt.Printf("%c %v\n", r, size)
	//
	// 	data = data[:len(data)-size]
	// }
}

func main4() {
	c := exec.Command("cat")
	// c.Args = []string{"-c cat"}
	winsize := pty.Winsize{Cols: 80, Rows: 24}
	ptmx, err := pty.StartWithSize(c, &winsize)
	if err != nil {
		log.Fatalln(err)
	}
	// term.MakeRaw(int(ptmx.Fd()))

	data, err := hex.DecodeString("5573696e672064656661756c74207461673a206c61746573740a6c61746573743a2050756c6c696e672066726f6d206c6962726172792f616c70696e650a0a1b5b31411b5b324b0d3830316266616136336566323a2050756c6c696e67206673206c61796572200d1b5b31421b5b31411b5b324b0d3830316266616136336566323a20446f776e6c6f6164696e67202032392e31376b422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a20446f776e6c6f6164696e672020312e3839394d422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a20566572696679696e6720436865636b73756d200d1b5b31421b5b31411b5b324b0d3830316266616136336566323a20446f776e6c6f616420636f6d706c657465200d1b5b31421b5b31411b5b324b0d3830316266616136336566323a2045787472616374696e67202033322e37376b422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a2045787472616374696e672020312e3730344d422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a2045787472616374696e672020322e3739394d422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a2050756c6c20636f6d706c657465200d1b5b31424469676573743a207368613235363a336337343937626630633761663933343238323432643631373665386637393035663232303164386663353836316634356265376133343662356632333433360a5374617475733a20446f776e6c6f61646564206e6577657220696d61676520666f7220616c70696e653a6c61746573740a646f636b65722e696f2f6c6962726172792f616c70696e653a6c61746573740a")

	l, err := ptmx.Write(data)
	log.Println("write", l, err)
	buf := make([]byte, 1024)
	for {
		l, err := ptmx.Read(buf)
		log.Println("read", l, err, string(buf[:l]))
		time.Sleep(100 * time.Millisecond)
	}

}
func main3() {
	data, err := hex.DecodeString("5573696e672064656661756c74207461673a206c61746573740a6c61746573743a2050756c6c696e672066726f6d206c6962726172792f616c70696e650a0a1b5b31411b5b324b0d3830316266616136336566323a2050756c6c696e67206673206c61796572200d1b5b31421b5b31411b5b324b0d3830316266616136336566323a20446f776e6c6f6164696e67202032392e31376b422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a20446f776e6c6f6164696e672020312e3839394d422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a20566572696679696e6720436865636b73756d200d1b5b31421b5b31411b5b324b0d3830316266616136336566323a20446f776e6c6f616420636f6d706c657465200d1b5b31421b5b31411b5b324b0d3830316266616136336566323a2045787472616374696e67202033322e37376b422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a2045787472616374696e672020312e3730344d422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a2045787472616374696e672020322e3739394d422f322e3739394d420d1b5b31421b5b31411b5b324b0d3830316266616136336566323a2050756c6c20636f6d706c657465200d1b5b31424469676573743a207368613235363a336337343937626630633761663933343238323432643631373665386637393035663232303164386663353836316634356265376133343662356632333433360a5374617475733a20446f776e6c6f61646564206e6577657220696d61676520666f7220616c70696e653a6c61746573740a646f636b65722e696f2f6c6962726172792f616c70696e653a6c61746573740a")

	c := exec.Command("sh")
	c.Args = []string{"-c", "cat"}
	stdin, err := c.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Start(); err != nil {
		log.Fatal(err)
	}

	go func() {
		stdin.Write(data)
		stdin.Close()
	}()

	buf := make([]byte, 1024)
	for {
		l, err := stdout.Read(buf)
		log.Println("read", l, err, string(buf[:l]))
		time.Sleep(100 * time.Millisecond)
	}

}
*/
