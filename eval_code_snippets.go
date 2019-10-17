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

func evalCodeSnippetsForPage(page *Page) {
	book := page.Book
	// can happen for draft pages
	if book == nil {
		return
	}
	fn := func(block *notionapi.Block) {
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

func evalCodeSnippets(book *Book) {
	for _, page := range book.idToPage {
		evalCodeSnippetsForPage(page)
	}
}
