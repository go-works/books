package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
)

func mdfmtFile(path string) error {
	path = filepath.Clean(path)
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	d2 := fixupCodeBlocks(d)
	if !bytes.Equal(d, d2) {
		if err = ioutil.WriteFile(path, d2, 0644); err != nil {
			return err
		}
	}

	script := filepath.Join("mdfmt", "mdfmt.js")
	cmd := exec.Command("node", script, path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		s := strings.Join(cmd.Args, " ")
		fmt.Printf("[%s] failed with error '%s'. Output:\n%s\n", s, err, string(out))
		return err
	}
	return ioutil.WriteFile(path, out, 0644)
}
