package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	flgOpenBrowser bool
	flgHTTPAddr    string
	flgDev         bool
	flgDataDir     string

	dataDirCached = ""
)

func getDataDirMust() string {
	if dataDirCached != "" {
		return dataDirCached
	}
	// TODO: a better way to signal this
	// maybe -data-dir cmd-line flag
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		// assume we're running in dev
		dir, err := os.UserHomeDir()
		must(err)
		dir = filepath.Join(dir, "data", "myservice")
		err = os.MkdirAll(dir, 0755)
		must(err)
		dataDirCached = dir
		return dataDirCached
	}

	// assume running in production, on linux server
	dir := "/data/myservice"
	dataDirCached = makeDirMust(dir)
	return dataDirCached
}

func getHTTPLogDirMust() string {
	dir := filepath.Join(getDataDirMust(), "log_http")
	return makeDirMust(dir)
}

func main() {
	fmt.Printf("Hello\n")
}
