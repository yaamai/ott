package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	//	"strings"
	"flag"
)

var (
	testFileName string
	logLevelStr  string
)

func main() {
	// parse flags
	flag.StringVar(&logLevelStr, "log", "debug", "log level")
	flag.Parse()

	if flag.NArg() < 1 {
		os.Exit(1)
	}
	testFileName = flag.Args()[0]

	// initialize logging
	logConfig := zap.NewDevelopmentConfig()
	logLevel := new(zapcore.Level)
	logLevel.UnmarshalText([]byte(logLevelStr))
	logConfig.Level.SetLevel(*logLevel)

	logger, err := logConfig.Build()
	if err != nil {
		log.Fatalln(err)
	}
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	// parse t file
	f, err := os.OpenFile(testFileName, os.O_RDONLY, 0755)
	if err != nil {
		zap.S().Fatal(err)
		os.Exit(1)
	}
	defer f.Close()

	lines, err := ParseRawT(f)
	if err != nil {
		zap.S().Fatal(err)
		os.Exit(1)
	}

	// run
	testFile := NewFromRawT(lines)
	runner, err := NewRunner()
	if err != nil {
		zap.S().Fatal(err)
		os.Exit(1)
	}
	runner.Run(&testFile)
}
