package main

import (
	"log"
	"strings"
)

func main() {
	s, err := NewSession()
	if err != nil {
		log.Fatalln(err)
	}

	/*
	f, err := os.OpenFile("", os.O_RDONLY, 0755)
	if err != nil {
		return TFile{}, err
	}
	defer f.Close()
	*/
	stream := strings.NewReader(`
# meta
#  a: 100
#  b: 100
echo-a:
  $ echo -e "a\nb"
  a
  b
  $ echo -e "c\nd"
  c
`)
	t, err := ParseTFile(stream)
	if err != nil {
		log.Fatalln(err)
	}
	Run(s, t)
}
