package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	// top-level directory where .html files are generated
	destDir = "www"
	tmplDir = "tmpl"
)

const (
	gitHubBaseURL = "https://github.com/essentialbooks/books"
	notionBaseURL = "https://notion.so/"
	siteBaseURL   = "https://www.programming-books.io"
)

var (
	pathAppJS      = "/s/app.js"
	pathMainCSS    = "/s/main.css"
	pathIndexCSS   = "/s/index.css"
	pathFaviconICO = "/s/favicon.ico"
)

var (
	// directory where generated .html files for books are
	destEssentialDir = filepath.Join(destDir, "essential")

	templates *template.Template

	funcMap = template.FuncMap{
		"inc": tmplInc,
	}
)

func tmplInc(i int) int {
	return i + 1
}

func loadTemplatesMust() *template.Template {
	// we reload templates in preview mode
	if templates != nil && !flgPreviewOnDemand {
		return templates
	}
	pattern := filepath.Join("tmpl", "*.tmpl.html")
	var err error
	templates, err = template.New("").Funcs(funcMap).ParseGlob(pattern)
	//templates, err = template.ParseGlob(pattern)
	must(err)
	templates.Funcs(funcMap)
	return templates
}

func execTemplateToFileSilentMaybeMust(name string, data interface{}, path string) error {
	var errToReturn error
	tmpl := loadTemplatesMust()
	if tmpl == nil {
		return nil
	}
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, name, data)
	maybePanicIfErr(err)

	d := buf.Bytes()
	if doMinify {
		d2, err := minifier.Bytes("text/html", d)
		//maybePanicIfErr(err)
		if err == nil {
			totalHTMLBytes += len(d)
			totalHTMLBytesMinified += len(d2)
			d = d2
		} else {
			errToReturn = err
		}
	}
	err = ioutil.WriteFile(path, d, 0644)
	maybePanicIfErr(err)
	return errToReturn
}

func execTemplateToFileMaybeMust(name string, data interface{}, path string) error {
	return execTemplateToFileSilentMaybeMust(name, data, path)
}

func execTemplateToWriter(name string, data interface{}, w io.Writer) error {
	tmpl := loadTemplatesMust()
	return tmpl.ExecuteTemplate(w, name, data)
}

// PageCommon is a common information for most pages
type PageCommon struct {
	Analytics      template.HTML
	PathAppJS      string
	PathMainCSS    string
	PathIndexCSS   string
	PathFaviconICO string
}

func getPageCommon() PageCommon {
	return PageCommon{
		Analytics:      googleAnalytics,
		PathAppJS:      pathAppJS,
		PathMainCSS:    pathMainCSS,
		PathIndexCSS:   pathIndexCSS,
		PathFaviconICO: pathFaviconICO,
	}
}

func gen404TopLevel() {
	d := struct {
		PageCommon
		Book *Book
	}{
		PageCommon: getPageCommon(),
	}
	path := filepath.Join(destDir, "404.html")
	_ = execTemplateToFileMaybeMust("404.tmpl.html", d, path)
}

func splitBooks(books []*Book) ([]*Book, []*Book) {
	var left []*Book
	var right []*Book
	for i, book := range books {
		if i%2 == 0 {
			left = append(left, book)
		} else {
			right = append(right, book)
		}
	}
	return left, right
}

func execTemplate(tmplName string, d interface{}, path string, w io.Writer) error {
	// this code path is for the preview on demand server
	if w != nil {
		return execTemplateToWriter(tmplName, d, w)
	}

	// this code path is for generating static files
	_ = execTemplateToFileMaybeMust(tmplName, d, path)
	return nil
}

func genIndex(books []*Book, w io.Writer) error {
	leftBooks, rightBooks := splitBooks(books)
	d := struct {
		PageCommon
		Books      []*Book
		LeftBooks  []*Book
		RightBooks []*Book
		NotionURL  string
	}{
		PageCommon: getPageCommon(),
		Books:      books,
		LeftBooks:  leftBooks,
		RightBooks: rightBooks,
		NotionURL:  gitHubBaseURL,
	}

	path := filepath.Join(destDir, "index.html")
	return execTemplate("index2.tmpl.html", d, path, w)
}

func genIndexGrid(books []*Book, w io.Writer) error {
	d := struct {
		PageCommon
		Books []*Book
	}{
		PageCommon: getPageCommon(),
		Books:      books,
	}
	path := filepath.Join(destDir, "index-grid.html")
	return execTemplate("index-grid.tmpl.html", d, path, w)
}

func genFeedback(w io.Writer) error {
	path := filepath.Join(destDir, "feedback.html")
	d := struct {
		PageCommon
	}{
		PageCommon: getPageCommon(),
	}
	return execTemplate("feedback.tmpl.html", d, path, w)
}

func genAbout(w io.Writer) error {
	d := getPageCommon()
	path := filepath.Join(destDir, "about.html")
	return execTemplate("about.tmpl.html", d, path, w)
}

// TODO: consolidate chapter/article html
func genArticle(book *Book, page *Page, currChapNo int, currArticleNo int, w io.Writer) error {
	if w == nil {
		addSitemapURL(page.CanonnicalURL())
	}

	d := struct {
		PageCommon
		*Page
		CurrentChapterNo int
		CurrentArticleNo int
		Description      string
	}{
		PageCommon:       getPageCommon(),
		Page:             page,
		CurrentChapterNo: currChapNo,
		CurrentArticleNo: currArticleNo,
		Description:      page.Title,
	}

	path := page.destFilePath()
	err := execTemplate("article.tmpl.html", d, path, w)
	if err != nil {
		fmt.Printf("Failed to minify page %s in book %s\n", page.NotionID, book.Title)
	}
	return err
}

func genChapter(book *Book, page *Page, currNo int, w io.Writer) error {
	if w == nil {
		addSitemapURL(page.CanonnicalURL())
		for i, article := range page.Pages {
			_ = genArticle(book, article, currNo, i, nil)
		}
	}

	path := page.destFilePath()
	d := struct {
		PageCommon
		*Page
		CurrentChapterNo int
		Description      string
	}{
		PageCommon:       getPageCommon(),
		Page:             page,
		CurrentChapterNo: currNo,
		Description:      page.Title,
	}
	err := execTemplate("chapter.tmpl.html", d, path, w)
	if err != nil {
		return err
	}

	for _, imagePath := range page.images {
		imageName := filepath.Base(imagePath)
		dst := page.destImagePath(imageName)
		_ = copyFileMaybeMust(dst, imagePath)
	}
	return nil
}

func buildIDToPage(book *Book) {
	pages := book.GetAllPages()
	for _, page := range pages {
		id := toNoDashID(page.NotionPage.ID)
		book.idToPage[id] = page
		page.Book = book
	}
}

func bookPagesToHTML(book *Book) {
	nProcessed := 0
	pages := book.GetAllPages()
	for _, page := range pages {
		html := notionToHTML(page, book)
		page.BodyHTML = template.HTML(string(html))
		nProcessed++
	}
	fmt.Printf("bookPagesToHTML: processed %d pages for book %s\n", nProcessed, book.TitleLong)
}

func genBookIndex(book *Book, w io.Writer) error {
	data := struct {
		PageCommon
		Book *Book
	}{
		PageCommon: getPageCommon(),
		Book:       book,
	}

	path := filepath.Join(book.destDir(), "index.html")
	return execTemplate("book_index.tmpl.html", data, path, w)
}

func genBook404(book *Book, w io.Writer) error {
	data := struct {
		PageCommon
		Book *Book
	}{
		PageCommon: getPageCommon(),
		Book:       book,
	}

	path := filepath.Join(book.destDir(), "404.html")
	return execTemplate("404.tmpl.html", data, path, nil)
}

func genBook(book *Book) {
	log("Started genering book %s\n", book.Title)
	timeStart := time.Now()

	buildIDToPage(book)
	genContributorsPage(book)
	bookPagesToHTML(book)

	genBookTOCSearchMust(book)

	// generate index.html for the book
	err := os.MkdirAll(book.destDir(), 0755)
	maybePanicIfErr(err)
	if err != nil {
		return
	}

	_ = genBookIndex(book, nil)

	// TODO: per-book 404 should link to top of book, not top of website
	_ = genBook404(book, nil)

	addSitemapURL(book.CanonnicalURL())

	for i, chapter := range book.Chapters() {
		_ = genChapter(book, chapter, i, nil)
	}

	fmt.Printf("Generated book '%s' in %s\n", book.Title, time.Since(timeStart))
}
