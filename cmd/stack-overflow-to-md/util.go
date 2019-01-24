package main

import (
	"os"
	"path/filepath"

	"github.com/kjk/u"
)

func createDirForFileMust(path string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	u.PanicIfErr(err)
}
