package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
)

/*
Todo:
- move extractNotionIDFromURL to notionapi so that I can re-use it
- improve style of .img class (take from notionapi)
- set the right margin-bottom to .title
- notionapi: fix rendering of lists like in 9e2a7d8c43bb46f29962b0d2f195e19c
*/

// HTMLRenderer is for notion -> HTML generation
type HTMLRenderer struct {
	page *Page
	book *Book

	notionClient *notionapi.Client
	r            *tohtml.HTMLRenderer
}

// change https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
// =>
// url within the book
func (r *HTMLRenderer) maybeReplaceNotionLink(uri string) string {
	id := notionapi.ExtractNoDashIDFromNotionURL(uri)
	if id == "" {
		return uri
	}
	page := r.book.idToPage[id]
	if page == nil {
		lg("Didn't find page with id '%s' extracted from url %s\n", id, uri)
		return uri
	}
	page.Book = r.book
	return page.URL()
}

func (r *HTMLRenderer) getURLAndTitleForBlock(block *notionapi.Block) (string, string) {
	id := normalizeID(block.ID)
	page := r.book.idToPage[id]
	if page == nil {
		title := cleanTitle(block.Title)
		lg("No article for id %s %s\n", id, title)
		url := "/article/" + id + "/" + urlify(title)
		return url, title
	}

	return page.URL(), page.Title
}

func findPageByID(book *Book, id string) *Page {
	pages := book.GetAllPages()
	for _, page := range pages {
		if strings.EqualFold(page.getID(), id) {
			return page
		}
	}
	return nil
}

func (r *HTMLRenderer) reportIfInvalidLink(uri string) {
	link := r.maybeReplaceNotionLink(uri)
	if link != uri {
		return
	}
	if strings.HasPrefix(uri, "http") {
		return
	}
	pageID := normalizeID(r.page.getID())
	lg("Found invalid link '%s' in page https://notion.so/%s", uri, pageID)
	destPage := findPageByID(r.book, uri)
	if destPage != nil {
		lg(" most likely pointing to https://notion.so/%s\n", normalizeID(destPage.NotionPage.ID))
	} else {
		lg("\n")
	}
}

// renderInlineLink renders a link in inline block
// we replace inter-notion urls to inter-blog urls
func (r *HTMLRenderer) renderInlineLink(b *notionapi.InlineBlock) (string, bool) {
	r.reportIfInvalidLink(b.Link)
	link := r.maybeReplaceNotionLink(b.Link)
	text := html.EscapeString(b.Text)
	s := fmt.Sprintf(`<a href="%s">%s</a>`, link, text)
	return s, true
}

// RenderEmbed renders BlockEmbed
func (r *HTMLRenderer) RenderEmbed(block *notionapi.Block, entering bool) bool {
	uri := block.FormatEmbed.DisplaySource
	if strings.Contains(uri, "onlinetool.io/") {
		r.genGitEmbed(block)
		return true
	}
	if strings.Contains(uri, "repl.it/") {
		r.genReplitEmbed(block)
		return true
	}
	panicIf(true, "unsupported embed %s", uri)
	return false
}

func (r *HTMLRenderer) genReplitEmbed(block *notionapi.Block) {
	uri := block.FormatEmbed.DisplaySource
	uri = strings.Replace(uri, "?lite=true", "", -1)
	lg("Page: https://notion.so/%s\n", r.page.NotionID)
	lg("  Replit: %s\n", uri)
	panic("we no longer use replit")
}

func (r *HTMLRenderer) genSourceFile(sf *SourceFile) {
	{
		var tmp bytes.Buffer
		code := sf.CodeToShow()
		lang := sf.Lang
		htmlHighlight(&tmp, string(code), lang, "")
		d := tmp.Bytes()
		info := CodeBlockInfo{
			Lang:      sf.Lang,
			GitHubURI: sf.GitHubURL,
		}
		info.PlaygroundURI = sf.PlaygroundURI
		s := fixupHTMLCodeBlock(string(d), &info)
		r.r.WriteString(s)
	}

	output := sf.Output()
	if len(output) != 0 {
		var tmp bytes.Buffer
		htmlHighlight(&tmp, output, "text", "")
		d := tmp.Bytes()
		info := CodeBlockInfo{
			Lang: "output",
		}
		s := fixupHTMLCodeBlock(string(d), &info)
		r.r.WriteString(s)
	}
}

func (r *HTMLRenderer) genGitEmbed(block *notionapi.Block) {
	uri := block.FormatEmbed.DisplaySource
	f := findSourceFileForEmbedURL(r.page, uri)
	// currently we only handle source code file embeds but might handle
	// others (graphs etc.)
	if f == nil {
		lg("genEmbed: didn't find source file for url %s\n", uri)
		return
	}

	r.genSourceFile(f)

}

// TODO: compare this with notionapi
/*
func (r *HTMLRenderer) genCollectionView(block *notionapi.Block) {
	viewInfo := block.CollectionViews[0]
	view := viewInfo.CollectionView
	if view.Format == nil {
		lg("genCollectionView: missing view.Format block id: %s\n", block.ID)
		return
	}

	s := `<table class="notion-table">`

	columns := view.Format.TableProperties
	s += `<thead><tr>`
	for _, col := range columns {
		colName := col.Property
		colInfo := viewInfo.Collection.CollectionSchema[colName]
		name := ""
		if colInfo != nil {
			name = colInfo.Name
		} else {
			lg("Missing colInfo in block ID '%s', page: %s\n", block.ID, r.page.NotionID)
		}
		s += `<th>` + html.EscapeString(name) + `</th>`
	}
	s += `</tr></thead>`

	s += `<tbody>`
	for _, row := range viewInfo.CollectionRows {
		s += `<tr>`
		props := row.Properties
		for _, col := range columns {
			colName := col.Property
			v := props[colName]
			colVal := propsValueToText(v)
			if colVal == "" {
				// use &nbsp; so that empty row still shows up
				// could also set a min-height to 1em or sth. like that
				s += `<td>&nbsp;</td>`
			} else {
				//colInfo := viewInfo.Collection.CollectionSchema[colName]
				// TODO: format colVal according to colInfo
				s += `<td>` + html.EscapeString(colVal) + `</td>`
			}
		}
		s += `</tr>`
	}
	s += `</tbody>`
	s += `</table>`
	r.writeString(s)
}
*/

// TODO: compare this with notionapi
/*
// Children of BlockColumnList are BlockColumn blocks
func (r *HTMLRenderer) genColumnList(block *notionapi.Block) {
	panicIf(block.Type != notionapi.BlockColumnList, "unexpected block type '%s'", block.Type)
	nColumns := len(block.Content)
	panicIf(nColumns == 0, "has no columns")
	// TODO: for now equal width columns
	s := `<div class="column-list">`
	r.writeString(s)

	for _, col := range block.Content {
		// TODO: get column ration from col.FormatColumn.ColumnRation, which is float 0...1
		panicIf(col.Type != notionapi.BlockColumn, "unexpected block type '%s'", col.Type)
		r.writeString(`<div>`)
		r.genBlocks(col.Content)
		r.writeString(`</div>`)
	}

	s = `</div>`
	r.writeString(s)
}
*/

func (r *HTMLRenderer) getInlineContent(block *notionapi.Block) string {
	r.r.PushNewBuffer()
	r.r.RenderInlines(block.InlineContent)
	return r.r.PopBuffer().String()
}

// RenderHeaders renders headers
// TODO: re-use code from notionapi
func (r *HTMLRenderer) RenderHeaders(block *notionapi.Block, entering bool) bool {
	tag := ""
	switch block.Type {
	case notionapi.BlockHeader:
		tag = "h1"
	case notionapi.BlockSubHeader:
		tag = "h2"
	case notionapi.BlockSubSubHeader:
		tag = "h3"
	default:
		panic("unsupported block type")
	}

	// avoid expensive work when exiting
	if !entering {
		r.r.WriteElement(block, tag, nil, "", entering)
		return true
	}

	id := notionapi.ToNoDashID(block.ID)
	h := HeadingInfo{
		Text: r.getInlineContent(block),
		ID:   id,
	}
	r.page.Headings = append(r.page.Headings, h)
	attrs := []string{"class", "hdr"}
	r.r.WriteElement(block, tag, attrs, "", entering)
	return true
}

// RenderCode renders BlockCode
func (r *HTMLRenderer) RenderCode(block *notionapi.Block, entering bool) bool {
	if !entering {
		return true
	}
	//lang := getLangFromFileExt(filepath.Ext(path))
	//gitHubURL := getGitHubPathForFile(path)
	lang := block.CodeLanguage
	sf := &SourceFile{
		NotionOriginURL: fmt.Sprintf("https://notion.so/%s", normalizeID(r.page.NotionID)),
		//Path:      path,
		//FileName:  name,
		//GitHubURL: gitHubURL,
	}
	sf.Lang = lang
	sf.SnippetName = r.page.PageTitle()
	if sf.SnippetName == "" {
		sf.SnippetName = "untitled"
	}

	data := []byte(block.Code)
	err := setSourceFileData(sf, data)
	if err != nil {
		lg("genBlock: setSourceFileData() failed with '%s'\n", err)
		lg("page: %s\n", sf.NotionOriginURL)
		//panicIfErr(err)
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
	err = getOutputCached(r.book.cache, sf)
	if err != nil {
		lg("getOutputCached() failed.\nsf.CodeToRun():\n%s\n", sf.CodeToRun)
		panicIfErr(err)
	}
	r.genSourceFile(sf)

	if false {
		// code := template.HTMLEscapeString(block.Code)
		//fmt.Fprintf(g.f, `<div class="%s">Lang for code: %s</div>
		//<pre class="%s">
		//%s
		//</pre>`, levelCls, block.CodeLanguage, levelCls, code)
		var tmp bytes.Buffer
		htmlHighlight(&tmp, string(block.Code), block.CodeLanguage, "")
		d := tmp.Bytes()
		var info CodeBlockInfo
		// TODO: set Lang, GitHubURI and PlaygroundURI
		s := fixupHTMLCodeBlock(string(d), &info)
		r.r.WriteString(s)
	}
	return true
}

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
		fmt.Printf("detectFileNameFromLanguage: lang '%s' is not supported\n", sf.Lang)
		fmt.Printf("Notion page: %s\n", sf.NotionOriginURL)
		panic("")
	}
	sf.Directive.FileName = "main" + ext
	if sf.FileName == "" {
		sf.FileName = sf.Directive.FileName
		sf.Path = sf.FileName
	}
	return nil
}

// RenderImage renders BlockImage
// TODO: download images locally like blog
func (r *HTMLRenderer) RenderImage(block *notionapi.Block, entering bool) bool {
	link := block.ImageURL
	cls := "img"
	attrs := []string{"class", cls, "src", link}
	r.r.WriteElement(block, "img", attrs, "", entering)
	return true
}

// RenderPage renders BlockPage
func (r *HTMLRenderer) RenderPage(block *notionapi.Block, entering bool) bool {
	/*
		cls := "page"
		if block.IsLinkToPage() {
			cls = "page-link"
		}
		url, title := r.getURLAndTitleForBlock(block)
		title = template.HTMLEscapeString(title)
		html := fmt.Sprintf(`<div class="%s%s"><a href="%s">%s</a></div>`, cls, levelCls, url, title)
		fmt.Fprintf(r.f, "%s\n", html)
	*/
	tp := block.GetPageType()
	if tp == notionapi.BlockPageTopLevel {
		// skips top-level as it's rendered somewhere else
		return true
	}

	var cls string
	if tp == notionapi.BlockPageSubPage {
		cls = "page"
	} else if tp == notionapi.BlockPageLink {
		cls = "page-link"
	} else {
		panic("unexpected page type")
	}

	url, title := r.getURLAndTitleForBlock(block)
	title = html.EscapeString(title)
	content := fmt.Sprintf(`<a href="%s">%s</a>`, url, title)
	attrs := []string{"class", cls}
	title = template.HTMLEscapeString(title)
	r.r.WriteElement(block, "div", attrs, content, entering)
	return true
}

var (
	toggleEntering = `
<div style="width: 100%%; margin-top: 2px; margin-bottom: 1px;">
    <div style="display: flex; align-items: flex-start; width: 100%%; padding-left: 2px; color: rgb(66, 66, 65);">

        <div style="margin-right: 4px; width: 24px; flex-grow: 0; flex-shrink: 0; display: flex; align-items: center; justify-content: center; min-height: calc((1.5em + 3px) + 3px); padding-right: 2px;">
            <div id="toggle-toggle-{{id}}" onclick="javascript:onToggleClick(this)" class="toggler" style="align-items: center; user-select: none; display: flex; width: 1.25rem; height: 1.25rem; justify-content: center; flex-shrink: 0;">

                <svg id="toggle-closer-{{id}}" width="100%%" height="100%%" viewBox="0 0 100 100" style="fill: currentcolor; display: none; width: 0.6875em; height: 0.6875em; transition: transform 300ms ease-in-out; transform: rotateZ(180deg);">
                    <polygon points="5.9,88.2 50,11.8 94.1,88.2 "></polygon>
                </svg>

                <svg id="toggle-opener-{{id}}" width="100%%" height="100%%" viewBox="0 0 100 100" style="fill: currentcolor; display: block; width: 0.6875em; height: 0.6875em; transition: transform 300ms ease-in-out; transform: rotateZ(90deg);">
                    <polygon points="5.9,88.2 50,11.8 94.1,88.2 "></polygon>
                </svg>
            </div>
        </div>

        <div style="flex: 1 1 0px; min-width: 1px;">
            <div style="display: flex;">
                <div style="padding-top: 3px; padding-bottom: 3px">{{inline}}</div>
            </div>

            <div style="margin-left: -2px; display: none" id="toggle-content-{{id}}">
                <div style="display: flex; flex-direction: column;">
                    <div style="width: 100%%; margin-top: 2px; margin-bottom: 0px;">
                        <div style="color: rgb(66, 66, 65);">
							<div style="">
`
	toggleClosing = `
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>
`
)

// In notion I want to have @TODO lines that are not rendered in html output
func isBlockTextTodo(block *notionapi.Block) bool {
	panicIf(block.Type != notionapi.BlockText, "only supported on '%s' block, called on '%s' block", notionapi.BlockText, block.Type)
	blocks := block.InlineContent
	if len(blocks) == 0 {
		return false
	}
	b := blocks[0]
	if strings.HasPrefix(b.Text, "@TODO") {
		return true
	}
	return false
}

func isBlockTextEmpty(block *notionapi.Block) bool {
	panicIf(block.Type != notionapi.BlockText, "only supported on '%s' block, called on '%s' block", notionapi.BlockText, block.Type)
	blocks := block.InlineContent
	if len(blocks) == 0 {
		return true
	}
	return false
}

func (r *HTMLRenderer) isLastBlock() bool {
	lastIdx := len(r.r.CurrBlocks) - 1
	return r.r.CurrBlockIdx == lastIdx
}

func (r *HTMLRenderer) isFirstBlock() bool {
	return r.r.CurrBlockIdx == 0
}

// RenderText renders BlockText
func (r *HTMLRenderer) RenderText(block *notionapi.Block, entering bool) bool {
	if isBlockTextTodo(block) {
		return true
	}

	// notionapi/tohtml renders empty blocks as visible, so skip empty text
	// blocks if they are the first or last. Assumption is that it's careless
	// editing
	skipIfEmpty := r.isLastBlock() || r.isFirstBlock()
	if skipIfEmpty && isBlockTextEmpty(block) {
		return true
	}

	// TODO: convert to div
	r.r.WriteElement(block, "p", nil, "", entering)
	return true
}

// RenderToggle renders BlockToggle blocks
func (r *HTMLRenderer) RenderToggle(block *notionapi.Block, entering bool) bool {
	panicIf(block.Type != notionapi.BlockToggle, "unexpected block type '%s'", block.Type)

	if entering {
		// TODO: could do it without pushing buffers
		r.r.PushNewBuffer()
		r.r.RenderInlines(block.InlineContent)
		inline := r.r.PopBuffer().String()
		id := notionapi.ToNoDashID(block.ID)
		s := strings.Replace(toggleEntering, "{{id}}", id, -1)
		s = strings.Replace(s, "{{inline}}", inline, -1)
		r.r.WriteString(s)

	} else {
		r.r.WriteString(toggleClosing)
	}
	// we handled it
	return true
}

func (r *HTMLRenderer) blockRenderOverride(block *notionapi.Block, entering bool) bool {
	switch block.Type {
	case notionapi.BlockPage:
		return r.RenderPage(block, entering)
	case notionapi.BlockCode:
		return r.RenderCode(block, entering)
	case notionapi.BlockToggle:
		return r.RenderToggle(block, entering)
	case notionapi.BlockImage:
		return r.RenderImage(block, entering)
	case notionapi.BlockText:
		return r.RenderText(block, entering)
	case notionapi.BlockEmbed:
		return r.RenderEmbed(block, entering)
	case notionapi.BlockHeader, notionapi.BlockSubHeader, notionapi.BlockSubSubHeader:
		return r.RenderHeaders(block, entering)
	}
	return false
}

// Gen returns generated HTML
func (r *HTMLRenderer) Gen() []byte {
	inner := string(r.r.ToHTML())

	rootPage := r.page.NotionPage.Root
	f := rootPage.FormatPage
	isMono := f != nil && f.PageFont == "mono"

	s := ``
	if isMono {
		s += `<div style="font-family: monospace">`
	}
	s += inner
	if isMono {
		s += `</div>`
	}
	return []byte(s)
}

func notionToHTML(page *Page, book *Book) []byte {
	// This is artificially generated page (e.g. contributors page)
	if page.NotionPage == nil {
		return []byte(page.BodyHTML)
	}

	verbose("Generating HTML for %s\n", page.NotionURL())
	res := HTMLRenderer{
		book: book,
		page: page,
	}

	r := tohtml.NewHTMLRenderer(page.NotionPage)
	r.PanicOnFailures = true
	r.AddIDAttribute = true
	r.Data = res
	r.RenderBlockOverride = res.blockRenderOverride
	r.RenderInlineLinkOverride = res.renderInlineLink

	res.r = r

	return res.Gen()
}
