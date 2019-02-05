package main

import (
	"strings"

	"github.com/essentialbooks/books/pkg/common"
)

func isMaybeCodeBlock(s string) bool {
	return strings.HasPrefix(s, "    ")
}

func addBuffered(a []string, buffered []string) []string {
	if len(buffered) == 0 {
		return a
	}
	if len(buffered) == 1 {
		return append(a, buffered[0])
	}
	a = append(a, "```")
	for _, s := range buffered {
		s = s[4:]
		a = append(a, s)
	}
	a = append(a, "```")
	return a
}

/*
converts code blocks from indented form to fenced form.
*/
func fixupCodeBlocks(d []byte) []byte {
	d = common.NormalizeNewlines(d)
	lines := strings.Split(string(d), "\n")
	var newLines []string
	var buffered []string
	for _, line := range lines {
		if isMaybeCodeBlock(line) {
			buffered = append(buffered, line)
			continue
		}
		newLines = addBuffered(newLines, buffered)
		buffered = nil
		newLines = append(newLines, line)
	}
	s := strings.Join(newLines, "\n")
	return []byte(s)
}
