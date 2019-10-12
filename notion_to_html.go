package main

import (
	"bytes"
	"html"
	"strings"

	"github.com/essentialbooks/books/tohtml"
	"github.com/kjk/notionapi"

	//"github.com/kjk/notionapi/tohtml"
	"github.com/kjk/u"
)

/*
Todo:
- improve style of .img class (take from notionapi)
- set the right margin-bottom to .title
*/

// Converter is for notion -> HTML generation
type Converter struct {
	page *Page
	book *Book

	notionClient *notionapi.Client
	converter    *tohtml.Converter
}

func areNotionIDsEqual(id1, id2 string) bool {
	id1 = toNoDashID(id1)
	id2 = toNoDashID(id2)
	return id1 == id2
}

func (c *Converter) reportIfInvalidLink(uri string, extractedID string) {
	pageID := c.page.getID()
	log("Found invalid link '%s' (id: '%s') in page https://notion.so/%s\n", uri, extractedID, pageID)
	if extractedID == "" {
		return
	}
	page := findPageByID(c.book, extractedID)
	if page != nil {
		log(" strange, we actually found it via findPageByID()\n")
	}
}

// change https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
// =>
// url within the book
func (c *Converter) rewriteURL(uri string) string {
	if !strings.Contains(uri, "notion.so/") {
		return uri
	}

	id := notionapi.ExtractNoDashIDFromNotionURL(uri)
	if id == "" {
		c.reportIfInvalidLink(uri, id)
		return uri
	}
	page := c.book.idToPage[id]
	if page == nil {
		c.reportIfInvalidLink(uri, id)
		return uri
	}
	page.Book = c.book
	return page.URL()
}

func (c *Converter) getURLAndTitleForBlock(block *notionapi.Block) (string, string) {
	id := toNoDashID(block.ID)
	page := c.book.idToPage[id]
	if page == nil {
		title := cleanTitle(block.Title)
		log("No article for id %s %s\n", id, title)
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

// RenderEmbed renders BlockEmbed
func (c *Converter) RenderEmbed(block *notionapi.Block) bool {
	uri := block.FormatEmbed().DisplaySource
	if strings.Contains(uri, "onlinetool.io/") {
		c.genGitEmbed(block)
		return true
	}
	if strings.Contains(uri, "repl.it/") {
		c.genReplitEmbed(block)
		return true
	}
	u.PanicIf(true, "unsupported embed %s", uri)
	return false
}

func (c *Converter) genReplitEmbed(block *notionapi.Block) {
	uri := block.FormatEmbed().DisplaySource
	uri = strings.Replace(uri, "?lite=true", "", -1)
	log("Page: https://notion.so/%s\n", c.page.NotionID)
	log("  Replit: %s\n", uri)
	panic("we no longer use replit")
}

func (c *Converter) genSourceFile(sf *SourceFile) {
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
		c.converter.Printf(s)
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
		c.converter.Printf(s)
	}
}

func (c *Converter) genGitEmbed(block *notionapi.Block) {
	uri := block.FormatEmbed().DisplaySource
	f := findSourceFileForEmbedURL(c.page, uri)
	// currently we only handle source code file embeds but might handle
	// others (graphs etc.)
	if f == nil {
		log("genEmbed: didn't find source file for url %s\n", uri)
		return
	}

	c.genSourceFile(f)
}

// RenderCode renders BlockCode
func (c *Converter) RenderCode(block *notionapi.Block) bool {

	sf := c.page.blockCodeToSourceFile[block.ID]
	c.genSourceFile(sf)

	if false {
		// code := html.EscapeString(block.Code)
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
		c.converter.Printf(s)
	}
	return true
}

// RenderImage renders BlockImage
// TODO: download images locally like blog
func (c *Converter) RenderImage(block *notionapi.Block) bool {
	link := block.ImageURL
	c.converter.Printf(`<img class="img" src="%s">`, link)
	return true
}

// RenderPage renders BlockPage
func (c *Converter) RenderPage(block *notionapi.Block) bool {
	if c.converter.Page.IsRoot(block) {
		// skips top-level as it's rendered somewhere else
		c.converter.RenderChildren(block)
		return true
	}

	cls := "page-link"
	if block.IsSubPage() {
		cls = "page"
	}

	url, title := c.getURLAndTitleForBlock(block)
	title = html.EscapeString(title)
	c.converter.Printf(`<div class="%s">
<a href="%s">%s</a>`, cls, url, title)
	c.converter.RenderChildren(block)
	c.converter.Printf("`</div>")
	return true
}

// In notion I want to have @TODO lines that are not rendered in html output
func isBlockTextTodo(block *notionapi.Block) bool {
	u.PanicIf(block.Type != notionapi.BlockText, "only supported on '%s' block, called on '%s' block", notionapi.BlockText, block.Type)
	blocks := block.InlineContent
	if len(blocks) == 0 {
		return false
	}
	b := blocks[0]
	switch {
	case strings.HasPrefix(b.Text, "@TODO"):
		return true
	case strings.HasPrefix(b.Text, "#TODO"):
		return true
	}
	return false
}

func isBlockTextEmpty(block *notionapi.Block) bool {
	u.PanicIf(block.Type != notionapi.BlockText, "only supported on '%s' block, called on '%s' block", notionapi.BlockText, block.Type)
	blocks := block.InlineContent
	if len(blocks) == 0 {
		return true
	}
	return false
}

func (c *Converter) isLastBlock() bool {
	lastIdx := len(c.converter.CurrBlocks) - 1
	return c.converter.CurrBlockIdx == lastIdx
}

func (c *Converter) isFirstBlock() bool {
	return c.converter.CurrBlockIdx == 0
}

// RenderText renders BlockText
func (c *Converter) RenderText(block *notionapi.Block) bool {
	if isBlockTextTodo(block) {
		return true
	}

	// notionapi/tohtml renders empty blocks as visible, so skip empty text
	// blocks if they are the first or last. Assumption is that it's careless
	// editing
	skipIfEmpty := c.isLastBlock() || c.isFirstBlock()
	if skipIfEmpty && isBlockTextEmpty(block) {
		return true
	}

	// TODO: convert to div
	c.converter.Printf(`<p>`)
	c.converter.RenderInlines(block.InlineContent)
	c.converter.RenderChildren(block)
	c.converter.Printf(`</p>`)
	return true
}

func (c *Converter) blockRenderOverride(block *notionapi.Block) bool {
	switch block.Type {
	case notionapi.BlockPage:
		return c.RenderPage(block)
	case notionapi.BlockCode:
		return c.RenderCode(block)
	case notionapi.BlockImage:
		return c.RenderImage(block)
	case notionapi.BlockText:
		return c.RenderText(block)
	case notionapi.BlockEmbed:
		return c.RenderEmbed(block)
	}
	return false
}

// Gen returns generated HTML
func (c *Converter) Gen() []byte {
	/*
		// using notionapi/tohtml converter
		inner, err := c.converter.ToHTML()
		must(err)
	*/

	// using our tohtml converter
	inner := c.converter.ToHTML()

	rootPage := c.page.NotionPage.Root()
	f := rootPage.FormatPage()
	isMono := f != nil && f.PageFont == "mono"

	s := ``
	if isMono {
		s += `<div style="font-family: monospace">`
	}
	s += string(inner)
	if isMono {
		s += `</div>`
	}
	return []byte(s)
}

func getInlinesPlain(a []*notionapi.TextSpan) string {
	s := ""
	for _, b := range a {
		s += b.Text
	}
	return s
}

func notionToHTML(page *Page, book *Book) []byte {
	// This is artificially generated page (e.g. contributors page)
	if page.NotionPage == nil {
		return []byte(page.BodyHTML)
	}

	logVerbose("Generating HTML for %s\n", page.NotionURL())
	res := Converter{
		book: book,
		page: page,
	}

	r := tohtml.NewConverter(page.NotionPage)
	notionapi.PanicOnFailures = true
	r.RenderBlockOverride = res.blockRenderOverride
	r.RewriteURL = res.rewriteURL
	res.converter = r

	html := res.Gen()

	return html
}
