package main

import "os"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func makeDirMust(dir string) string {
	err := os.MkdirAll(dir, 0755)
	must(err)
	return dir
}
