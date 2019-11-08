package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kjk/u"
)

var allowedLanguages = map[string]bool{
	"go":         true,
	"javascript": true,
	"cpp":        true,
}

func updateGlotID(cache *Cache, sf *SourceFile) error {
	u.PanicIf(!sf.Directive.Glot)

	// we don't want it
	if sf.Directive.NoPlayground {
		return nil
	}

	// already have it
	if sf.GlotPlaygroundID != "" {
		return nil
	}

	// it's already in cache
	sha1 := sf.Sha1()
	id := cache.sha1ToGlotID[sha1]
	if id != "" {
		sf.GlotPlaygroundID = id
		sf.PlaygroundURI = "https://glot.io/snippets/" + sf.GlotPlaygroundID
		return nil
	}

	lang := strings.ToLower(sf.Lang)
	lang = glotConvertLanguage(lang)
	if _, ok := allowedLanguages[lang]; !ok {
		return fmt.Errorf("'%s' ('%s') is not a supported language", sf.Lang, lang)
	}

	fileName := sf.Directive.FileName
	snippetName := sf.SnippetName

	d := []byte(sf.CodeToRun)
	rsp, err := glotGetSnippedID(d, snippetName, fileName, lang)
	if err != nil {
		return err
	}
	id = rsp.ID
	sf.GlotPlaygroundID = id
	sf.PlaygroundURI = "https://glot.io/snippets/" + sf.GlotPlaygroundID

	cache.saveGlotID(sha1, id)
	return nil
}

func updateGlotOutput(cache *Cache, sf *SourceFile) error {
	// we already have it
	if sf.GlotOutput != "" {
		return nil
	}

	// it's already in cache
	o := cache.sha1ToGlotOutput[sf.Sha1()]
	if o != nil {
		sf.GlotOutput = o.Output
		return nil
	}

	f := &glotFile{
		Name:    sf.Directive.FileName,
		Content: sf.CodeToRun,
	}
	if sf.Directive.RunCmd != "" {
		logf("  run command: %s\n", sf.Directive.RunCmd)
	}
	req := &glotRunRequest{
		Command:  sf.Directive.RunCmd,
		language: sf.Lang,
		Files:    []*glotFile{f},
	}
	rsp, err := glotRun(req)
	if err != nil {
		logf("glotRun() failed. Page: %s\n", sf.NotionOriginURL)
		u.Must(err)
	}
	s := rsp.Stdout + rsp.Stderr
	if rsp.Error != "" {
		if !sf.Directive.AllowError {
			logf("glotRun() %s from %s failed with '%s' and sf.Directive.AllowError is false\n", sf.SnippetName, sf.NotionOriginURL, rsp.Error)
			return errors.New(rsp.Stderr)
		}
	}
	sf.GlotOutput = s
	logf("Got glot output (%d bytes) for %s from %s\n", len(sf.GlotOutput), sf.SnippetName, sf.NotionOriginURL)

	o = &EvalOutput{
		Lang:      sf.Lang,
		FileName:  f.Name,
		CodeFull:  sf.CodeFull,
		CodeToRun: sf.CodeToRun,
		RunCmd:    sf.Directive.RunCmd,
		Output:    s,
	}
	if !flgNoUpdateOutput {
		cache.saveGlotOutput(o)
	}
	return nil
}

// for a given file, get output of executing this command
// We cache this as it is the most expensive part of rebuilding books
// If allowError is true, we silence an error from executed command
// This is useful when e.g. executing "go run" on a program that is
// intentionally not valid.
func getOutputCached(cache *Cache, sf *SourceFile) error {
	// TODO: make the check for when not to execute the file
	// even better
	if sf.Directive.NoOutput {
		if sf.Directive.Glot {
			// if running on glot, we want to execute even if
			// we don't show the output (to check syntax errors)
		} else {
			return nil
		}
	}

	if sf.Directive.Glot {
		err := updateGlotOutput(cache, sf)
		must(err)
		err = updateGlotID(cache, sf)
		must(err)
	}
	return nil
}
