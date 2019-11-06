package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kjk/siser"
	"github.com/kjk/u"
)

const (
	recNameGlotOutput = "glotoutput"
	recNameGlotID     = "glotid"
	recNameGist       = "gist"       // content of the gist
	recNameGistOutput = "gistoutput" // result of evaluating the gist
)

type Cache struct {
	// path of the cache file
	path string

	sha1ToGlotOutput map[string]*EvalOutput
	// id of glot.id code snippet for this code, if exists
	sha1ToGlotID map[string]string
	// id of the gist to gist content. Might be outdated if
	// gist changes
	gistIDToGist         map[string]string
	gistSha1ToGistOutput map[string]string
}

func NewCache(path string) *Cache {
	return &Cache{
		path:                 path,
		sha1ToGlotOutput:     map[string]*EvalOutput{},
		sha1ToGlotID:         map[string]string{},
		gistIDToGist:         map[string]string{},
		gistSha1ToGistOutput: map[string]string{},
	}
}

func (c *Cache) saveRecord(rec *siser.Record) error {
	f := openForAppend(c.path)
	defer u.CloseNoError(f)
	w := siser.NewWriter(f)
	_, err := w.WriteRecord(rec)
	return err
}

type EvalOutput struct {
	Lang      string
	FileName  string
	CodeFull  string
	CodeToRun string
	RunCmd    string
	Output    string

	// sha1 of CodeFull, calculated on demand
	sha1 string
}

func recGetMust(rec *siser.Record, name string) string {
	v, ok := rec.Get(name)
	u.PanicIf(!ok)
	return v
}

func recGetMustNonEmpty(rec *siser.Record, name string) string {
	v, ok := rec.Get(name)
	u.PanicIf(!ok || v == "")
	return v
}

func (c *EvalOutput) Sha1() string {
	if c.sha1 == "" {
		c.sha1 = u.Sha1HexOfBytes([]byte(c.CodeFull))
	}
	return c.sha1
}

func (c *Cache) saveGlotOutput(code *EvalOutput) {
	u.PanicIf(code.CodeFull == "")
	u.PanicIf(c.sha1ToGlotOutput[code.Sha1()] != nil)
	rec := &siser.Record{
		Name: recNameGlotOutput,
	}
	rec.Append("Sha1", code.Sha1())
	rec.Append("Lang", code.Lang)
	rec.Append("FileName", code.FileName)
	rec.Append("CodeFull", code.CodeFull)
	rec.Append("CodeToRun", code.CodeToRun)
	if code.RunCmd != "" {
		rec.Append("RunCmd", code.RunCmd)
	}
	rec.Append("Output", code.Output)
	err := c.saveRecord(rec)
	must(err)
	c.sha1ToGlotOutput[code.Sha1()] = code
}

func (c *Cache) loadGlotOutput(rec *siser.Record) {
	u.PanicIf(rec.Name != recNameGlotOutput)
	sha1, ok := rec.Get("Sha1")
	u.PanicIf(!ok || sha1 == "")
	u.PanicIf(c.sha1ToGlotOutput[sha1] != nil)

	o := &EvalOutput{}
	o.Lang, ok = rec.Get("Lang")
	u.PanicIf(!ok)
	o.FileName, ok = rec.Get("FileName")
	u.PanicIf(!ok)
	o.CodeFull, ok = rec.Get("CodeFull")
	u.PanicIf(!ok)
	o.CodeToRun, ok = rec.Get("CodeToRun")
	u.PanicIf(!ok)
	o.RunCmd, ok = rec.Get("RunCmd")
	o.Output, ok = rec.Get("Output")
	u.PanicIf(!ok)

	u.PanicIf(o.CodeFull == "")
	u.PanicIf(sha1 != o.Sha1(), "sha1 != code.Sha1() (%s != %s)", sha1, o.Sha1())
	c.sha1ToGlotOutput[sha1] = o
}

func (c *Cache) saveGlotID(sha1, glotID string) {
	u.PanicIf(c.sha1ToGlotID[sha1] != "")
	rec := &siser.Record{
		Name: recNameGlotID,
	}
	u.PanicIf(sha1 == "")
	rec.Append("Sha1", sha1)
	rec.Append("GlotID", glotID)
	err := c.saveRecord(rec)
	must(err)
	c.sha1ToGlotID[sha1] = glotID
}

func (c *Cache) loadGlotID(rec *siser.Record) {
	u.PanicIf(rec.Name != recNameGlotID)
	sha1 := recGetMustNonEmpty(rec, "Sha1")
	u.PanicIf(c.sha1ToGlotID[sha1] != "")

	glotid := recGetMustNonEmpty(rec, "GlotID")
	c.sha1ToGlotID[sha1] = glotid
}

func (c *Cache) saveGist(gistID, gist string) {
	rec := &siser.Record{
		Name: recNameGist,
	}
	u.PanicIf(gistID == "" || gist == "")
	rec.Append("GistID", gistID)
	rec.Append("Gist", gist)
	err := c.saveRecord(rec)
	must(err)
	c.gistIDToGist[gistID] = gist
}

func (c *Cache) loadGist(rec *siser.Record) {
	u.PanicIf(rec.Name != recNameGist)
	gistID := recGetMustNonEmpty(rec, "GistID")
	gist := recGetMustNonEmpty(rec, "Gist")
	c.gistIDToGist[gistID] = gist
}

func (c *Cache) saveGistOutput(gist, output string) {
	rec := &siser.Record{
		Name: recNameGistOutput,
	}
	//TODO: probably remove, it's ok to have no output
	//u.PanicIf(output == "")
	rec.Append("Gist", gist)
	rec.Append("GistOutput", output)
	err := c.saveRecord(rec)
	must(err)
	sha1 := u.Sha1HexOfBytes([]byte(gist))
	c.gistSha1ToGistOutput[sha1] = output
}

func (c *Cache) loadGistOutput(rec *siser.Record) {
	gist := recGetMustNonEmpty(rec, "Gist")
	output := recGetMust(rec, "GistOutput")
	sha1 := u.Sha1HexOfBytes([]byte(gist))
	c.gistSha1ToGistOutput[sha1] = output
}

func loadCache(path string) *Cache {
	logf("loadCache: %s\n", path)
	dir := filepath.Dir(path)
	// the directory must exist
	_, err := os.Stat(dir)
	must(err)

	c := NewCache(path)

	f, err := os.Open(path)
	if err != nil {
		// it's ok if file doesn't exist
		logf("  cache file %s doesn't exist\n", path)
		return c
	}
	defer u.CloseNoError(f)

	nRecords := 0
	r := siser.NewReader(bufio.NewReader(f))
	for r.ReadNextRecord() {
		rec := r.Record
		switch rec.Name {
		case recNameGlotOutput:
			c.loadGlotOutput(rec)
		case recNameGlotID:
			c.loadGlotID(rec)
		case recNameGist:
			c.loadGist(rec)
		case recNameGistOutput:
			c.loadGistOutput(rec)
		default:
			panic(fmt.Errorf("unknown record type: '%s'", rec.Name))
		}
		nRecords++
	}
	must(r.Err())
	logf(" got %d cache records\n", nRecords)
	return c
}
