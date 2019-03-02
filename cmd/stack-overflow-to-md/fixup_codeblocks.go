package main

import (
	"strings"

	"github.com/essentialbooks/books/pkg/common"
)

func isMaybeCodeBlock(s string, buffered []string) bool {
	if strings.HasPrefix(s, "    ") {
		return true
	}
	// a single empty line in the middle
	n := len(buffered)
	if n == 0 {
		return false
	}
	lastWasEmpty := buffered[n-1] == ""
	return s == "" && !lastWasEmpty
}

func calcIndent(s string) int {
	n := 0
	for i := range s {
		if s[i] == ' ' {
			n++
		} else {
			break
		}
	}
	return n
}

func addBuffered(a []string, buffered []string) []string {
	n := len(buffered)
	if n == 0 {
		return a
	}
	if n == 1 {
		// TODO: should this be wrapped in? Those happen.
		return append(a, buffered[0])
	}
	lastWasEmpty := buffered[n-1] == ""
	if lastWasEmpty {
		if n == 2 {
			// single code line followed by single empty line
			return append(a, buffered...)
		}
		buffered = buffered[:n-1]
	}

	firstIndent := calcIndent(buffered[0])
	a = append(a, "```")
	for _, s := range buffered {
		indent := calcIndent(s)
		if indent > firstIndent {
			indent = firstIndent
		}
		s = s[indent:]
		a = append(a, s)
	}
	a = append(a, "```")
	if lastWasEmpty {
		a = append(a, "")
	}
	return a
}

func isCodeBlockDelimiter(s string) bool {
	return strings.HasPrefix(s, "```")
}

/*
converts code blocks from indented form to fenced form.
*/
func fixupCodeBlocks(d []byte) []byte {
	d = common.NormalizeNewlines(d)
	lines := strings.Split(string(d), "\n")
	var newLines []string
	var buffered []string
	inCodeBlock := false
	for _, line := range lines {
		if isCodeBlockDelimiter(line) {
			inCodeBlock = !inCodeBlock
		}

		if isMaybeCodeBlock(line, buffered) && !inCodeBlock {
			buffered = append(buffered, line)
			continue
		}
		newLines = addBuffered(newLines, buffered)
		buffered = nil
		newLines = append(newLines, line)
	}
	newLines = addBuffered(newLines, buffered)
	s := strings.Join(newLines, "\n")
	return []byte(s)
}
