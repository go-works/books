package main

import (
	"fmt"

	"github.com/kjk/wc"
)

var srcFiles = wc.MakeAllowedFileFilterForExts(".go", ".js", ".html", ".css")
var excludeDirs = wc.MakeExcludeDirsFilter("books", ".vscode", ".github", ".idea", "cache", "code", "covers", "log", "mdfmt", "node_modules", "pkg", "essential", "s")
var allFiles = wc.MakeFilterAnd(srcFiles, excludeDirs)

func doLineCount() int {
	stats := wc.NewLineStats()
	err := stats.CalcInDir(".", allFiles, true)
	if err != nil {
		fmt.Printf("doLineCount: stats.wcInDir() failed with '%s'\n", err)
		return 1
	}
	wc.PrintLineStats(stats)
	return 0
}
