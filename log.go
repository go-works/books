package main

import (
	"fmt"
	"os"
)

var (
	logFile *os.File
)

func openLog() {
	var err error
	logFile, err = os.Create("log.txt")
	must(err)
}

func closeLog() {
	if logFile == nil {
		return
	}
	logFile.Close()
	logFile = nil
}

func log(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if logFile != nil {
		fmt.Fprint(logFile, s)
	}
	fmt.Print(s)
}

func logVerbose(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if logFile != nil {
		fmt.Fprint(logFile, s)
	}
}

func logFatal(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if logFile != nil {
		fmt.Fprint(logFile, s)
	}
	fmt.Print(s)
	os.Exit(1)
}
