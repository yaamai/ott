package main

import (
	"go.uber.org/zap"
	"log"
	"strings"
)

func main() {
	logConfig := zap.NewDevelopmentConfig()
	logConfig.Level.SetLevel(zap.InfoLevel)
	logger, err := logConfig.Build()
	if err != nil {
		log.Fatalln(err)
	}
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

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
  aaaa
  d
date:
  $ date
  aaa
multiline:
  $ export B=200
  $ echo a &&\
  > date &&\
  > echo $B
  b
`)
	t, err := ParseTFile(stream)
	if err != nil {
		log.Fatalln(err)
	}
	Run(s, t)
}
