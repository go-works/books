package main

import (
	"fmt"
	"os"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func logf(format string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Print(format)
		return
	}
	fmt.Printf(format, args...)
}

func makeDirMust(dir string) string {
	err := os.MkdirAll(dir, 0755)
	must(err)
	return dir
}
