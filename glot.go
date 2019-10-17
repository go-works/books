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

	"github.com/kjk/u"
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

func glotConvertLanguage(s string) string {
	s = strings.ToLower(s)
	switch s {
	case "c++", "cplusplus":
		return "cpp"
	}
	return s
}

func glotGetLangs() []*glotLang {
	if gGlotLangs != nil {
		return gGlotLangs
	}
	err := json.Unmarshal([]byte(glotLangsJSON), &gGlotLangs)
	u.Must(err)
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
	Command string `json:"command,omitempty"`
	// Stdin is optional
	Stdin string      `json:"stdin,omitempty"`
	Files []*glotFile `json:"files"`
}

type glotRunResponse struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error  string `json:"error"`
}

type glotMakeSnippetRequest struct {
	Language string      `json:"language"`
	Title    string      `json:"title"`
	Public   bool        `json:"public"`
	Files    []*glotFile `json:"files"`
}

type glotMakeSnippetResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
	// created
	// modified
	FilesHash string      `json:"files_hash"`
	Language  string      `json:"language"`
	Title     string      `json:"title"`
	Public    bool        `json:"public"`
	Owner     string      `json:"owner"`
	Files     []*glotFile `json:"files"`
}

func glotHTTPPostJSON(uri string, reqIn interface{}, rspOut interface{}) error {
	js, err := json.Marshal(reqIn)
	if err != nil {
		return err
	}
	hc := &http.Client{
		Timeout: 60 * time.Second,
	}
	body := bytes.NewBuffer(js)
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token 10f32ff1-fa02-4a7c-bb71-22d4ed5d286b")
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		d, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Request was '%s' (%d) and not OK (200). Body:\n%s\nurl: %s", resp.Status, resp.StatusCode, string(d), uri)
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logf("glotHTTPPostJSON: ioutil.ReadAll() failed with '%s'\n", err)
		return err
	}

	// TODO: this sucks because I can't tell a difference between
	// a valid "no output" (the program doesn't print anything)
	// or caused by an error (e.g. I forgot to name the files so nothing
	// gets executed)
	// special case: when doing 'run' requests and it returns 204 No Content
	// it might be a valid program that doesn't print anything to stdout
	// in which case trying to decode empty string as JSON will fail
	if resp.StatusCode == http.StatusNoContent {
		logf("glotHTTPPostJSON: got 204 No Content\n")
		if _, ok := reqIn.(*glotRunRequest); ok {
			if req, ok := reqIn.(*glotRunRequest); ok {
				for _, f := range req.Files {
					logf("Name: %s\n", f.Name)
					logf("Content:\n%s\n", f.Content)
				}
			}
			// allow Unmarshal to work (and set everything to default empty fields)
			d = []byte("{}")
		}
	}

	err = json.Unmarshal(d, rspOut)
	if err != nil {
		if req, ok := reqIn.(*glotRunRequest); ok {
			for _, f := range req.Files {
				logf("Name: %s\n", f.Name)
				logf("Content:\n%s\n", f.Content)
			}
		}

		logf("glotHTTPPostJSON: json.Unmarshal() failed with '%s' on:\n%s\n", err, string(d))
		logf("   uri: '%s', input type: %T, output type: %T\n", uri, reqIn, rspOut)
		logf("   status code: %d, status: '%s'\n", resp.StatusCode, resp.Status)
		return err
	}
	return nil
}

func glotRun(req *glotRunRequest) (*glotRunResponse, error) {
	req.language = glotConvertLanguage(req.language)
	runURL, err := glotFindRunURLForLang(req.language)
	if err != nil {
		return nil, err
	}
	var res *glotRunResponse
	err = glotHTTPPostJSON(runURL, req, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// submit the data to Glot playground and get snippet id
func glotGetSnippedID(content []byte, snippetName string, fileName string, lang string) (*glotMakeSnippetResponse, error) {
	if fileName == "" {
		panic("fileName cannot be empty")
	}
	lang = glotConvertLanguage(lang)
	_, err := glotFindRunURLForLang(lang)
	if err != nil {
		return nil, err
	}
	uri := "https://snippets.glot.io/snippets"

	file := &glotFile{
		Name:    fileName,
		Content: string(content),
	}
	req := &glotMakeSnippetRequest{
		Public:   true,
		Title:    snippetName,
		Language: lang,
		Files:    []*glotFile{file},
	}
	var res *glotMakeSnippetResponse
	err = glotHTTPPostJSON(uri, req, &res)
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

func glotGetSnippedIDTestAndExit() {
	s := `package main

import (
	"fmt"
	"strconv"
)

func main() {
	// :show start
	var i1 int = -38
	fmt.Printf("i1: %s\n", strconv.Itoa(i1))

	var i2 int32 = 148
	fmt.Printf("i2: %s\n", strconv.Itoa(int(i2)))
	// :show end
}
	`
	res, err := glotGetSnippedID([]byte(s), "foo", "foo.go", "go")
	u.Must(err)
	fmt.Printf("share id: '%s', url: %s\n", res.ID, res.URL)
	os.Exit(0)
}
