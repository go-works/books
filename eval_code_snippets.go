package main

import (
	"fmt"
	"strings"

	"github.com/kjk/u"

	"github.com/kjk/notionapi"
)

func setDefaultFileNameFromLanguage(sf *SourceFile) error {
	if sf.Directive.FileName != "" {
		return nil
	}

	// we don't care unless it goes to glot.io
	if !sf.Directive.Glot {
		return nil
	}

	ext := ""
	lang := strings.ToLower(sf.Lang)
	switch lang {
	case "go":
		ext = ".go"
	case "javascript":
		ext = ".js"
	case "cpp", "cplusplus", "c++":
		ext = ".cpp"
	default:
		logf("detectFileNameFromLanguage: lang '%s' is not supported\n", sf.Lang)
		logf("Notion page: %s\n", sf.NotionOriginURL)
		panic("")
	}
	sf.Directive.FileName = "main" + ext
	if sf.FileName == "" {
		sf.FileName = sf.Directive.FileName
		sf.Path = sf.FileName
	}
	return nil
}

type CodeEvalInfo struct {
	URL    string
	GistID string
	// TODO: maybe more info, like which file to show if more than one
	// alternatively this might come from codeeval.yml
}

// returns empty string if s is not code eval url
func parseCodeEvalGist(s string) string {
	if !strings.Contains(s, "https://codeeval.dev/gist/") {
		return ""
	}
	gist := strings.TrimPrefix(s, "https://codeeval.dev/gist/")
	panicIf(len(gist) != len("5648d36550f7592236327e920243f791"), "'%s' doesn't look like a valid gist in url '%s'", gist, s)
	return gist
}

// parses a line like:
// https://codeeval.dev/gist/5648d36550f7592236327e920243f791 main.go
// return nil if is not code eval line
func parseCodeEvalInfo(s string) *CodeEvalInfo {
	if !strings.Contains(s, "https://codeeval.dev/gist/") {
		return nil
	}
	res := &CodeEvalInfo{}
	parts := strings.Split(s, " ")
	for _, uri := range parts {
		gistID := parseCodeEvalGist(uri)
		if gistID == "" {
			continue
		}
		panicIf(res.GistID != "", "has second uri: '%s', first: '%s'", uri, res.URL)
		res.GistID = gistID
		res.URL = s
	}
	return res
}

func evalCodeSnippetsForPage(page *Page) {
	book := page.Book
	// can happen for draft pages
	if book == nil {
		return
	}
	fnText := func(block *notionapi.Block) {
		panicIf(block.Type != notionapi.BlockText)
		ts := block.GetTitle()
		s := notionapi.TextSpansToString(ts)
		if strings.Contains(s, "https://codeeval.dev") {
			//logf("found it: %s\n", s)
			//os.Exit(1)
			// TODO: eval it
		}
	}

	fnEmbed := func(block *notionapi.Block) {
		panicIf(block.Type != notionapi.BlockEmbed)
		uri := block.FormatEmbed().DisplaySource
		if strings.Contains(uri, "https://codeeval.dev") {
			logf("found codeeval: '%s'\n", uri)
			panic("embed blocks NYI")
		}
	}

	fn := func(block *notionapi.Block) {
		if block.Type == notionapi.BlockText {
			fnText(block)
			return
		}

		if block.Type == notionapi.BlockEmbed {
			fnEmbed(block)
			return
		}

		if block.Type != notionapi.BlockCode {
			return
		}

		if page.blockCodeToSourceFile == nil {
			page.blockCodeToSourceFile = map[string]*SourceFile{}
		}

		if false {
			s := block.Code
			lines := dataToLines([]byte(s))
			firstLine := ""
			if len(lines) > 0 {
				firstLine = lines[0]
			}
			logf("Page: %s\n  %s\n", page.Title, firstLine)
		}

		//lang := getLangFromFileExt(filepath.Ext(path))
		//gitHubURL := getGitHubPathForFile(path)
		lang := block.CodeLanguage
		sf := &SourceFile{
			NotionOriginURL: fmt.Sprintf("https://notion.so/%s", toNoDashID(page.NotionID)),
			//Path:      path,
			//FileName:  name,
			//GitHubURL: gitHubURL,
		}
		sf.Lang = lang
		sf.SnippetName = page.PageTitle()
		if sf.SnippetName == "" {
			sf.SnippetName = "untitled"
		}

		data := []byte(block.Code)
		err := setSourceFileData(sf, data)
		if err != nil {
			logf("genBlock: setSourceFileData() failed with '%s'\n", err)
			logf("page: %s\n", sf.NotionOriginURL)
			//u.Must(err)
		}

		// TODO: this uses codeeval evaluator
		if false && sf.Directive.Glot {
			evalSourceFile(sf)
		}

		if sf.Directive.Glot || sf.Directive.GoPlayground {
			// for those we respect no output/no playground
		} else {
			// for embedded code blocks by default we don't set playground
			// or output unless explicitly asked for
			sf.Directive.NoPlayground = true
			sf.Directive.NoOutput = true
		}
		setDefaultFileNameFromLanguage(sf)
		err = getOutputCached(book.cache, sf)
		if err != nil {
			logf("getOutputCached() failed.\nsf.CodeToRun():\n%s\n", sf.CodeToRun)
			u.Must(err)
		}
		page.blockCodeToSourceFile[block.ID] = sf
	}

	page.NotionPage.ForEachBlock(fn)
}

/*
func evalCodeSnippets(book *Book) {
	for _, page := range book.idToPage {
		evalCodeSnippetsForPage(page)
	}
}
*/
