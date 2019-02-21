package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"runtime"
	"strings"
	"time"

	"html/template"

	"github.com/essentialbooks/books/pkg/common"
	"github.com/kjk/notionapi"
	"github.com/kjk/u"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

var (
	flgAnalytics      string
	flgPreview        bool
	flgNoCache        bool
	flgRecreateOutput bool
	flgUpdateOutput   bool
	flgRedownloadReplit bool
	flgRedownloadOne string
	flgRedownloadOneReplit string

	soUserIDToNameMap map[int]string
	googleAnalytics   template.HTML
	doMinify bool
	minifier *minify.M
)

var (
	bookGo = 		&Book{
		Title:     "Go",
		TitleLong: "Essential Go",
		Dir:       "go",
		CoverImageName: "Go.png",
		// https://www.notion.so/2cab1ed2b7a44584b56b0d3ca9b80185
		NotionStartPageID: "2cab1ed2b7a44584b56b0d3ca9b80185",
	}
	bookCsharp = 		&Book{
		Title: "C#",
		TitleLong: "Essential C#",
		Dir: "csharp",
		CoverImageName: "CSharp.png",
		// https://www.notion.so/kjkpublic/Essential-C-896da5248e65414ab645dd45985879a1
		NotionStartPageID: "896da5248e65414ab645dd45985879a1",
	}
	bookPython = &Book{
		Title: "Python",
		TitleLong: "Essential Python",
		Dir: "python",
		CoverImageName: "Python.png",
		// https://www.notion.so/kjkpublic/Essential-Python-12e6f78e68a5497290c96e1365ae6259
		NotionStartPageID: "12e6f78e68a5497290c96e1365ae6259",
	}
	bookKotlin = &Book{
		NoPublish: true,
		Title: "Kotlin",
		TitleLong: "Essential Kotlin",
		Dir: "kotlin",
		CoverImageName: "Bash.png", // TODO: need a cover
		// https://www.notion.so/kjkpublic/Essential-Kotlin-2bdd47318f3a4e8681dda289a8b3472b
		NotionStartPageID: "2bdd47318f3a4e8681dda289a8b3472b",
	}
	bookJavaScript = &Book{
		NoPublish: true,
		Title: "JavaScript",
		TitleLong: "Essential JavaScript",
		Dir: "javascript",
		CoverImageName: "JavaScript.png",
		// https://www.notion.so/kjkpublic/Essential-Javascript-0b121710a160402fa9fd4646b87bed99
		NotionStartPageID: "0b121710a160402fa9fd4646b87bed99",
	}
	bookDart = &Book{
		NoPublish: true,
		Title: "Dart",
		TitleLong: "Essential Dart",
		Dir: "dart",
		CoverImageName: "Cpp.png", // TODO: need a cover
		// 	https://www.notion.so/kjkpublic/Essential-Dart-0e2d248bf94b4aebaefbcf51ae435df0
		NotionStartPageID: "0e2d248bf94b4aebaefbcf51ae435df0",
	}
	bookJava = &Book{
		NoPublish: true,
		Title: "Java",
		TitleLong: "Essential Java",
		Dir: "java",
		CoverImageName: "Java.png",
		// https://www.notion.so/kjkpublic/Essential-Java-d37cda98a07046f6b2cc375731ea3bdb
		NotionStartPageID: "d37cda98a07046f6b2cc375731ea3bdb",
	}
	bookAndroid = &Book{
		NoPublish: true,
		Title: "Android",
		TitleLong: "Essential Android",
		Dir: "android",
		CoverImageName: "Android.png",
		// https://www.notion.so/kjkpublic/Essential-Android-f90b0a6b648343e28dc5ed6e8f5c0780
		NotionStartPageID: "f90b0a6b648343e28dc5ed6e8f5c0780",
	}
	bookSql = &Book{
		NoPublish: true,
		Title: "SQL",
		TitleLong: "Essential SQL",
		Dir: "sql",
		CoverImageName: "SQL.png",
		// https://www.notion.so/kjkpublic/Essential-SQL-d1c8bb39bad4494e80abe28414c3d80e
		NotionStartPageID: "d1c8bb39bad4494e80abe28414c3d80e",
	}
)

var (
	books = []*Book{
		bookGo, bookCsharp, bookPython, bookKotlin, bookJavaScript,
		bookDart, bookJava, bookAndroid, bookSql,
	}
)

func parseFlags() {
	flag.StringVar(&flgAnalytics, "analytics", "", "google analytics code")
	flag.BoolVar(&flgPreview, "preview", false, "if true will start watching for file changes and re-build everything")
	flag.BoolVar(&flgRecreateOutput, "recreate-output", false, "if true, recreates ouput files in cache")
	flag.BoolVar(&flgUpdateOutput, "update-output", false, "if true, will update ouput files in cache")
	flag.BoolVar(&flgNoCache, "no-cache", false, "if true, disables cache for notion")
	flag.StringVar(&flgRedownloadOne, "redownload-one", "", "notion id of a page to re-download")
	flag.BoolVar(&flgRedownloadReplit, "redownload-replit", false, "if true, redownloads replits")
	flag.StringVar(&flgRedownloadOneReplit, "redownload-one-replit", "", "replit url and book to download")

	flag.Parse()

	if flgRedownloadOne != "" {
		flgRedownloadOne = extractNotionIDFromURL(flgRedownloadOne)
	}

	if flgAnalytics != "" {
		googleAnalyticsTmpl := `<script async src="https://www.googletagmanager.com/gtag/js?id=%s"></script>
		<script>
			window.dataLayer = window.dataLayer || [];
			function gtag(){dataLayer.push(arguments);}
			gtag('js', new Date());
			gtag('config', '%s')
		</script>
	`
		s := fmt.Sprintf(googleAnalyticsTmpl, flgAnalytics, flgAnalytics)
		googleAnalytics = template.HTML(s)
	}
}

func downloadBook(c *notionapi.Client, book *Book) {
	notionStartPageID := book.NotionStartPageID
	book.pageIDToPage = map[string]*notionapi.Page{}
	loadNotionPages(book, c, notionStartPageID, book.pageIDToPage, !flgNoCache)
	fmt.Printf("Loaded %d pages for book %s\n", len(book.pageIDToPage), book.Title)
	bookFromPages(book)
}

func loadSOUserMappingsMust() {
	path := filepath.Join("stack-overflow-docs-dump", "users.json.gz")
	err := common.JSONDecodeGzipped(path, &soUserIDToNameMap)
	u.PanicIfErr(err)
}

// TODO: probably more
func getDefaultLangForBook(bookName string) string {
	s := strings.ToLower(bookName)
	switch s {
	case "go":
		return "go"
	case "android":
		return "java"
	case "ios":
		return "ObjectiveC"
	case "microsoft sql server":
		return "sql"
	case "node.js":
		return "javascript"
	case "mysql":
		return "sql"
	case ".net framework":
		return "c#"
	}
	return s
}

func shouldCopyImage(path string) bool {
	return !strings.Contains(path, "@2x")
}

func copyCoversMust() {
	copyFilesRecur(filepath.Join("www", "covers"), "covers", shouldCopyImage)
}

func getAlmostMaxProcs() int {
	// leave some juice for other programs
	nProcs := runtime.NumCPU() - 2
	if nProcs < 1 {
		return 1
	}
	return nProcs
}

// copy from tmpl to www, optimize if possible, add
// sha1 of the content as part of the name
func copyToWwwAsSha1MaybeMust(srcName string) {
	var dstPtr *string
	minifyType := ""
	switch srcName {
	case "main.css":
		dstPtr = &pathMainCSS
		minifyType = "text/css"
	case "app.js":
		dstPtr = &pathAppJS
		minifyType = "text/javascript"
	case "favicon.ico":
		dstPtr = &pathFaviconICO
	default:
		panicIf(true, "unknown srcName '%s'", srcName)
	}
	src := filepath.Join("tmpl", srcName)
	d, err := ioutil.ReadFile(src)
	panicIfErr(err)

	if doMinify && minifyType != "" {
		d2, err := minifier.Bytes(minifyType, d)
		maybePanicIfErr(err)
		if err == nil {
			fmt.Printf("Compressed %s from %d => %d (saved %d)\n", srcName, len(d), len(d2), len(d)-len(d2))
			d = d2
		}
	}

	sha1Hex := u.Sha1HexOfBytes(d)
	name := nameToSha1Name(srcName, sha1Hex)
	dst := filepath.Join("www", "s", name)
	err = ioutil.WriteFile(dst, d, 0644)
	panicIfErr(err)
	*dstPtr = filepath.ToSlash(dst[len("www"):])
	fmt.Printf("Copied %s => %s\n", src, dst)
}

func genAllBooks() {
	nProcs := getAlmostMaxProcs()

	timeStart := time.Now()
	clearSitemapURLS()
	copyCoversMust()

	copyToWwwAsSha1MaybeMust("main.css")
	copyToWwwAsSha1MaybeMust("app.js")
	copyToWwwAsSha1MaybeMust("favicon.ico")
	genIndex(books)
	genIndexGrid(books)
	gen404TopLevel()
	genAbout()
	genFeedback()

	for _, book := range books {
		genBook(book)
	}
	writeSitemap()
	fmt.Printf("Used %d procs, finished generating all books in %s\n", nProcs, time.Since(timeStart))
}

func initMinify() {
	minifier = minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("text/html", html.Minify)
	minifier.AddFunc("text/javascript", js.Minify)
	// less aggresive minification because html validators
	// report this as html errors
	minifier.Add("text/html", &html.Minifier{
		KeepDocumentTags: true,
		KeepEndTags:      true,
	})
	doMinify = !flgPreview
}

func findBookFromDir(dir string) *Book {
	for _, book := range books {
		if book.Dir == dir {
			return book
		}
	}
	return nil
}

func isNotionCachedInDir(dir string, id string) bool {
	id = normalizeID(id)
	files, err := ioutil.ReadDir(dir)
	panicIfErr(err)
	for _, fi := range files {
		name := fi.Name()
		if strings.HasPrefix(name, id) {
			return true
		}
	}
	return false
}

func findBookFromCachedPageID(id string) *Book {
	files, err := ioutil.ReadDir("cache")
	panicIfErr(err)
	for _, book := range books {
		if book.NotionStartPageID == id {
			return book
		}
	}

	for _, fi := range files {
		if !fi.IsDir() {
			continue
		}
		dir := fi.Name()
		book := findBookFromDir(dir)
		panicIf(book == nil, "didn't find book for dir '%s'", dir)
		if isNotionCachedInDir(filepath.Join("cache", dir, "notion"), id) {
			return book
		}
	}
	return nil
}

func isReplitURL(uri string) bool {
	return strings.Contains(uri, "repl.it/")
}

func findBookByName(bookName string) *Book {
	for _, book := range books {
		if bookName == book.Dir {
			return book
		}
	}
	return nil
}

func redownloadOneReplit() {
	if len(flag.Args()) != 1 {
		fmt.Printf("-redownload-one-replit expects 2 arguments: book and replit url\n")
		os.Exit(1)
	}
	uri := flgRedownloadOneReplit
	bookName := flag.Args()[0]
	if !isReplitURL(uri) {
		panicIf(!isReplitURL(bookName), "neither '%s' nor '%s' look like repl.it url", uri, bookName)
		uri, bookName = bookName, uri
	}
	book := findBookByName(bookName)
	panicIf(book == nil, "'%s' is not a valid book name", bookName)
	initBook(book)
	_, isNew, err := downloadAndCacheReplit(book.replitCache, uri)
	panicIfErr(err)
	fmt.Printf("genReplitEmbed: downloaded %s,  isNew: %v\n", uri+".zip", isNew)
}

func initBook(book *Book) {
	var err error
	book.titleSafe = common.MakeURLSafe(book.Title)

	createDirMust(book.OutputCacheDir())
	createDirMust(book.NotionCacheDir())

	reloadCachedOutputFilesMust(book)
	path := filepath.Join(book.OutputCacheDir(), "sha1_to_go_playground_id.txt")
	book.sha1ToGoPlaygroundCache = readSha1ToGoPlaygroundCache(path)
	book.replitCache, err = LoadReplitCache(book.ReplitCachePath())
	panicIfErr(err)
}

func main() {
	parseFlags()

	if flgRedownloadOneReplit != "" {
		redownloadOneReplit()
		os.Exit(0)
	}

	if false {
		// only needs to be run when we add new covers
		genTwitterImagesAndExit()
	}

	os.RemoveAll("www")
	createDirMust(filepath.Join("www", "s"))
	createDirMust("log")

	client := &notionapi.Client{}
	if flgRedownloadOne != "" {
		book := findBookFromCachedPageID(flgRedownloadOne)
		if book == nil {
			fmt.Printf("didn't find book for id %s\n", flgRedownloadOne)
			os.Exit(1)
		}
		fmt.Printf("Downloading %s for book %s\n", flgRedownloadOne, book.Dir)
		// download a single page from notion and re-generate content
		_, err := downloadAndCachePage(book, client, flgRedownloadOne)
		if err != nil {
			fmt.Printf("downloadAndCachePage of '%s' failed with %s\n", flgRedownloadOne, err)
			os.Exit(1)
		}
		flgPreview = true
		// and fallthrough to re-generate books
	}

	initMinify()
	loadSOUserMappingsMust()

	if flgUpdateOutput {
		// TODO: must be done somewhere else
		if flgRecreateOutput {
			//os.RemoveAll(cachedOutputDir)
		}
	}

	for _, book := range books {
		initBook(book)
		downloadBook(client, book)
		loadSoContributorsMust(book)
	}

	genAllBooks()
	genNetlifyHeaders()
	genNetlifyRedirects()
	printAndClearErrors()

	if flgUpdateOutput || flgRedownloadOne != "" {
		for _, b := range books {
			saveCachedOutputFiles(b)
		}
	}

	if flgPreview {
		startPreview()
	}
}
