package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestParseDirective(t *testing.T) {
	s := `// :glot, :name main.go, :run go run -race main.go, :allow error`
	d, _, err := extractFileDirectives([]string{s})
	assert.NoError(t, err)
	assert.Equal(t, d.Glot, true)
	assert.Equal(t, d.FileName, "main.go")
	assert.Equal(t, d.RunCmd, "go run -race main.go")
	assert.True(t, d.AllowError)
}
