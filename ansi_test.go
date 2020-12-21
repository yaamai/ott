package main

import (
	"bytes"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func MustDecodeHexStringList(in []string) []byte {
	return MustDecodeHexString(strings.Join(in, ""))
}

func MustDecodeHexString(in string) []byte {
	data, _ := hex.DecodeString(in)
	return data
}

func TestTerminalAutoExpand(t *testing.T) {
	term := NewTerminal(0, 0)
	assert.NotNil(t, term)

	term.Write([]byte("A"))
	assert.Equal(t, []string{"A"}, term.StringLines())

	b := bytes.Repeat([]byte("B"), AUTO_EXPAND_COL+4)
	term.Write(b)
	assert.Equal(t, []string{"A" + string(b)}, term.StringLines())

	c := bytes.Repeat([]byte("C"), AUTO_EXPAND_COL)
	term.Write(bytes.Repeat(append(c, 0x0a), AUTO_EXPAND_ROW+4))
	expect := []string{"A" + string(b) + string(c)}
	for idx := 1; idx < AUTO_EXPAND_ROW+4; idx++ {
		expect = append(expect, string(c))
	}
	expect = append(expect, string(""))
	assert.Equal(t, expect, term.StringLines())
}

func TestTerminal(t *testing.T) {
	/*
		data, err := hex.DecodeString(strings.Join([]string{
		}, ""))
	*/

	tests := []struct {
		in  []byte
		out []string
	}{
		{
			[]byte{},
			[]string{""}, // NOTE: check this behavior is correct
		},
		{
			MustDecodeHexString("5573696e672064656661756c74207461673a206c6174657374"),
			[]string{"Using default tag: latest"},
		},
		{
			MustDecodeHexStringList([]string{
				"5573696e672064656661756c74207461673a206c61746573740a",
				"6c61746573743a2050756c6c696e672066726f6d206c6962726172792f616c70696e650a",
				"0a",
				"1b5b31411b5b324b0d3830316266616136336566323a2050756c6c696e67206673206c61796572200d",
			}),
			[]string{
				"Using default tag: latest",
				"latest: Pulling from library/alpine",
				"801bfaa63ef2: Pulling fs layer ",
			},
		},
		{
			MustDecodeHexStringList([]string{
				"5573696e672064656661756c74207461673a206c61746573740a",
				"6c61746573743a2050756c6c696e672066726f6d206c6962726172792f616c70696e650a",
				"0a",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2050756c6c696e67206673206c61796572200d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20446f776e6c6f6164696e67202032392e31376b422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20446f776e6c6f6164696e672020312e3839394d422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20566572696679696e6720436865636b73756d200d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20446f776e6c6f616420636f6d706c657465200d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2045787472616374696e67202033322e37376b422f322e3739394d420d",
			}),
			[]string{
				"Using default tag: latest",
				"latest: Pulling from library/alpine",
				"801bfaa63ef2: Extracting  32.77kB/2.799MB",
			},
		},
		{
			MustDecodeHexStringList([]string{
				"5573696e672064656661756c74207461673a206c61746573740a",
				"6c61746573743a2050756c6c696e672066726f6d206c6962726172792f616c70696e650a",
				"0a",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2050756c6c696e67206673206c61796572200d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20446f776e6c6f6164696e67202032392e31376b422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20446f776e6c6f6164696e672020312e3839394d422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20566572696679696e6720436865636b73756d200d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20446f776e6c6f616420636f6d706c657465200d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2045787472616374696e67202033322e37376b422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2045787472616374696e672020312e3730344d422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2045787472616374696e672020322e3739394d422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2050756c6c20636f6d706c657465200d",
				"1b5b3142",
			}),
			[]string{
				"Using default tag: latest",
				"latest: Pulling from library/alpine",
				"801bfaa63ef2: Pull complete ",
				"",
			},
		},
		{
			MustDecodeHexStringList([]string{
				"5573696e672064656661756c74207461673a206c61746573740a",
				"6c61746573743a2050756c6c696e672066726f6d206c6962726172792f616c70696e650a",
				"0a",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2050756c6c696e67206673206c61796572200d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20446f776e6c6f6164696e67202032392e31376b422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20446f776e6c6f6164696e672020312e3839394d422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20566572696679696e6720436865636b73756d200d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a20446f776e6c6f616420636f6d706c657465200d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2045787472616374696e67202033322e37376b422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2045787472616374696e672020312e3730344d422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2045787472616374696e672020322e3739394d422f322e3739394d420d",
				"1b5b3142",
				"1b5b3141",
				"1b5b324b0d",
				"3830316266616136336566323a2050756c6c20636f6d706c657465200d",
				"1b5b3142",
				"4469676573743a207368613235363a336337343937626630633761663933343238323432643631373665386637393035663232303164386663353836316634356265376133343662356632333433360a",
				"5374617475733a20446f776e6c6f61646564206e6577657220696d61676520666f7220616c70696e653a6c61746573740a",
				"646f636b65722e696f2f6c6962726172792f616c70696e653a6c61746573740a",
			}),
			[]string{
				"Using default tag: latest",
				"latest: Pulling from library/alpine",
				"801bfaa63ef2: Pull complete ",
				"Digest: sha256:3c7497bf0c7af93428242d6176e8f7905f2201d8fc5861f45be7a346b5f23436",
				"Status: Downloaded newer image for alpine:latest",
				"docker.io/library/alpine:latest",
				"",
			},
		},
	}
	for _, tt := range tests {
		term := NewTerminal(80, 24)
		assert.NotNil(t, term)

		term.Write(tt.in)
		assert.Equal(t, tt.out, term.StringLines())
	}
}
