package main

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	//	"strings"
	"encoding/json"
	"flag"
)

var (
	testFileNameList []string
	logLevelStr      string
	outputMode       string
	outputFormat     string
	sessionMode      string
	sessionCmd       string
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

func parseTFiles(filenames []string) ([][]Line, []*TestFile) {
	linesList := [][]Line{}
	testFileList := []*TestFile{}

	for _, filename := range filenames {
		// parse t file
		f, err := os.OpenFile(filename, os.O_RDONLY, 0755)
		if err != nil {
			zap.S().Fatal(err)
			os.Exit(1)
		}
		defer f.Close()

		lines, err := ParseT(f)
		if err != nil {
			zap.S().Fatal(err)
			os.Exit(1)
		}

		testFile := NewFromRawT(filename, lines)
		testFileList = append(testFileList, &testFile)
		linesList = append(linesList, lines)
	}

	return linesList, testFileList
}

func parseFlags() {
	// parse flags
	flag.StringVar(&logLevelStr, "log", "warn", "log level")
	flag.StringVar(&outputMode, "mode", "diff", "output mode (diff/actual/expected)")
	flag.StringVar(&outputFormat, "format", "text", "output format (text/json)")
	flag.StringVar(&sessionMode, "session-mode", "shell", "session parse mode (shell/python)")
	flag.StringVar(&sessionCmd, "session-cmd", "bash", "session command")
	flag.Parse()

	if flag.NArg() < 1 {
		os.Exit(1)
	}
	testFileNameList = flag.Args()
}

func main() {
	// init flags and logger
	parseFlags()
	undo := initLog(logLevelStr)
	defer undo()

	_, testFileList := parseTFiles(testFileNameList)

	runner, err := NewRunner(sessionCmd, sessionMode)
	if err != nil {
		zap.S().Fatal(err)
		os.Exit(1)
	}
	resultFileList := runner.RunMultiple(testFileList)

	// output
	if outputFormat == "json" {
		jsonBytes, err := json.Marshal(resultFileList)
		if err != nil {
			zap.S().Fatal(err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		zap.S().Info("Converting " + outputMode + "mode")
		for _, f := range resultFileList {
			for _, l := range f.ConvertToLines(outputMode) {
				fmt.Println(l.Line())
			}
		}
	}
}
