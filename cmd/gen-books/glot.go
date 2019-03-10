package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

// https://github.com/prasmussen/glot-run/blob/master/api_docs/run.md

var (

	// from https://run.glot.io/languages
	glotLangsJSON = `[
		{
		"name":"assembly",
		"url":"https://run.glot.io/languages/assembly"
		},
		{
		"name":"ats",
		"url":"https://run.glot.io/languages/ats"
		},
		{
		"name":"bash",
		"url":"https://run.glot.io/languages/bash"
		},
		{
		"name":"c",
		"url":"https://run.glot.io/languages/c"
		},
		{
		"name":"clojure",
		"url":"https://run.glot.io/languages/clojure"
		},
		{
		"name":"cobol",
		"url":"https://run.glot.io/languages/cobol"
		},
		{
		"name":"coffeescript",
		"url":"https://run.glot.io/languages/coffeescript"
		},
		{
		"name":"cpp",
		"url":"https://run.glot.io/languages/cpp"
		},
		{
		"name":"crystal",
		"url":"https://run.glot.io/languages/crystal"
		},
		{
		"name":"csharp",
		"url":"https://run.glot.io/languages/csharp"
		},
		{
		"name":"d",
		"url":"https://run.glot.io/languages/d"
		},
		{
		"name":"elixir",
		"url":"https://run.glot.io/languages/elixir"
		},
		{
		"name":"elm",
		"url":"https://run.glot.io/languages/elm"
		},
		{
		"name":"erlang",
		"url":"https://run.glot.io/languages/erlang"
		},
		{
		"name":"fsharp",
		"url":"https://run.glot.io/languages/fsharp"
		},
		{
		"name":"go",
		"url":"https://run.glot.io/languages/go"
		},
		{
		"name":"groovy",
		"url":"https://run.glot.io/languages/groovy"
		},
		{
		"name":"haskell",
		"url":"https://run.glot.io/languages/haskell"
		},
		{
		"name":"idris",
		"url":"https://run.glot.io/languages/idris"
		},
		{
		"name":"java",
		"url":"https://run.glot.io/languages/java"
		},
		{
		"name":"javascript",
		"url":"https://run.glot.io/languages/javascript"
		},
		{
		"name":"julia",
		"url":"https://run.glot.io/languages/julia"
		},
		{
		"name":"kotlin",
		"url":"https://run.glot.io/languages/kotlin"
		},
		{
		"name":"lua",
		"url":"https://run.glot.io/languages/lua"
		},
		{
		"name":"mercury",
		"url":"https://run.glot.io/languages/mercury"
		},
		{
		"name":"nim",
		"url":"https://run.glot.io/languages/nim"
		},
		{
		"name":"ocaml",
		"url":"https://run.glot.io/languages/ocaml"
		},
		{
		"name":"perl",
		"url":"https://run.glot.io/languages/perl"
		},
		{
		"name":"perl6",
		"url":"https://run.glot.io/languages/perl6"
		},
		{
		"name":"php",
		"url":"https://run.glot.io/languages/php"
		},
		{
		"name":"python",
		"url":"https://run.glot.io/languages/python"
		},
		{
		"name":"ruby",
		"url":"https://run.glot.io/languages/ruby"
		},
		{
		"name":"rust",
		"url":"https://run.glot.io/languages/rust"
		},
		{
		"name":"scala",
		"url":"https://run.glot.io/languages/scala"
		},
		{
		"name":"swift",
		"url":"https://run.glot.io/languages/swift"
		},
		{
		"name":"typescript",
		"url":"https://run.glot.io/languages/typescript"
		}
		]
`
)

type glotLang struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

var (
	gGlotLangs []*glotLang
)

func glotGetLangs() []*glotLang {
	if gGlotLangs != nil {
		return gGlotLangs
	}
	err := json.Unmarshal([]byte(glotLangsJSON), &gGlotLangs)
	panicIfErr(err)
	return gGlotLangs
}

func glotFindRunURLForLang(lang string) (string, error) {
	langs := glotGetLangs()
	for _, gl := range langs {
		if strings.EqualFold(lang, gl.Name) {
			return gl.URL + "/latest", nil
		}
	}
	return "", fmt.Errorf("'%s' is not recognized glot language", lang)
}

type glotFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type glotRunRequest struct {
	// should be one of recognized languages
	language string

	// Command is optional
	Command string      `json:"command,omitempty"`
	Files   []*glotFile `json:"files"`
}

type glotRunResponse struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error  string `json:"error"`
}

func glotHTTPPostJSON(uri string, json []byte) ([]byte, error) {
	hc := &http.Client{
		Timeout: 5 * time.Second,
	}
	body := bytes.NewBuffer(json)
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token 10f32ff1-fa02-4a7c-bb71-22d4ed5d286b")
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		d, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Request was '%s' (%d) and not OK (200). Body:\n%s\nurl: %s", resp.Status, resp.StatusCode, string(d), uri)
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func glotRun(req *glotRunRequest) (*glotRunResponse, error) {
	runURL, err := glotFindRunURLForLang(req.language)
	if err != nil {
		return nil, err
	}
	d, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data, err := glotHTTPPostJSON(runURL, d)
	if err != nil {
		return nil, err
	}
	var res *glotRunResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func glotRunTestAndExit() {
	f1 := &glotFile{
		Name:    "main.py",
		Content: "print(42)",
	}
	req := &glotRunRequest{
		language: "python",
		Files:    []*glotFile{f1},
	}
	resp, err := glotRun(req)
	if err != nil {
		fmt.Printf("glotRun() failed with '%s'\n", err)
		os.Exit(1)
	}
	fmt.Printf("glotRun() returned:\n")
	pretty.Print(resp)
	os.Exit(0)
}
