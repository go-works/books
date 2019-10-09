package main

import (
	"fmt"
	"os"

	"github.com/kjk/u"
)

var (
	logFile *os.File
)

func openLog() func() {
	var err error
	logFile, err = os.Create("log.txt")
	must(err)
	return func() {
		_ = logFile.Close()
		logFile = nil
	}
}

func log(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if logFile != nil {
		_, _ = fmt.Fprint(logFile, s)
	}
	fmt.Print(s)
}

func logVerbose(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if logFile != nil {
		_, _ = fmt.Fprint(logFile, s)
	}
}

func logFatal(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if logFile != nil {
		_, _ = fmt.Fprint(logFile, s)
	}
	fmt.Print(s)
	os.Exit(1)
}

func logf(format string, args ...interface{}) {
	u.Logf(format, args...)
}
