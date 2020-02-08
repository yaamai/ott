package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"fmt"
	"os"
	//	"strings"
	"flag"
    "encoding/json"
)

var (
	testFileName string
	logLevelStr  string
    outputMode string
    outputFormat string
)

func initLog(level string) func() {
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
    return undo
}

func main() {
	// parse flags
	flag.StringVar(&logLevelStr, "log", "warn", "log level")
	flag.StringVar(&outputMode, "mode", "diff", "output mode (diff/actual/expected)")
	flag.StringVar(&outputFormat, "format", "text", "output format (text/json)")
	flag.Parse()

	if flag.NArg() < 1 {
		os.Exit(1)
	}
	testFileName = flag.Args()[0]

    undo := initLog(logLevelStr)
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

    // output
    if outputFormat == "json" {
        jsonBytes, err := json.Marshal(testFile)
        if err != nil {
            zap.S().Fatal(err)
        }
        fmt.Println(string(jsonBytes))
    } else {
        zap.S().Info("Converting " + outputMode + "mode")
        for _, l := range(testFile.ConvertToLines(outputMode)) {
            fmt.Println(l.Line())
        }
    }
}
