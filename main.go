package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"html/template"

	"github.com/essentialbooks/books/pkg/common"
	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/caching_downloader"
	"github.com/kjk/u"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

var (
	flgAnalytics       string
	flgPreviewStatic   bool
	flgPreviewOnDemand bool
	flgAllBooks        bool
	// if true, disables downloading pages
	flgNoDownload bool
	// if true, disables notion cache, forcing re-download of notion page
	// even if cached verison on disk exits
	flgDisableNotionCache bool
	// url or id of the page to rebuild
	flgNoUpdateOutput bool

	flgReportExternalLinks      bool
	flgReportStackOverflowLinks bool

	soUserIDToNameMap map[int]string
	googleAnalytics   template.HTML
	doMinify          bool
	minifier          *minify.M

	notionAuthToken string

	// when downloading pages from the server, count total number of
	// downloaded and those from cache
	nTotalDownloaded int
	nTotalFromCache  int
)

var (
	booksMain = []*Book{
		bookGo,
		bookCpp,
		bookJavaScript,
		bookCSS,
		bookHTML,
		bookHTMLCanvas,
		bookJava,
		bookKotlin,
		bookCsharp,
		bookPython,
		bookPostgresql,
		bookMysql,
		bookIOS,
		bookAndroid,
		bookBash,
		bookPowershell,
		bookBatch,
		bookGit,
		bookPHP,
		bookRuby,
		bookNode,
		bookDart,
		bookTypeScript,
		bookSwift,
	}
	booksUnpublished = []*Book{
		bookNETFramework,
		bookAlgorithm,
		bookC,
		bookObjectiveC,
		bookReact,
		bookReactNative,
		bookRubyOnRails,
		bookSql,
	}
	allBooks = append(booksMain, booksUnpublished...)
)

func parseFlags() {
	flag.StringVar(&flgAnalytics, "analytics", "", "google analytics code")
	flag.BoolVar(&flgPreviewStatic, "preview-static", false, "if true starts web server for previewing locally generated static html")
	flag.BoolVar(&flgPreviewOnDemand, "preview-on-demand", false, "if true will start web server for previewing the book locally")
	flag.BoolVar(&flgAllBooks, "all-books", false, "if true will do all books")
	flag.BoolVar(&flgNoUpdateOutput, "no-update-output", false, "if true, will disable updating ouput files in cache")
	flag.BoolVar(&flgDisableNotionCache, "no-cache", false, "if true, disables cache for notion")
	flag.BoolVar(&flgNoDownload, "no-download", false, "if true, will not download pages from notion")
	flag.BoolVar(&flgReportExternalLinks, "report-external-links", false, "if true, shows external links for all pages")
	flag.BoolVar(&flgReportStackOverflowLinks, "report-so-links", false, "if true, shows links to stackoverflow.com")
	flag.Parse()

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

	notionAuthToken = os.Getenv("NOTION_TOKEN")
	if notionAuthToken != "" {
		fmt.Printf("NOTION_TOKEN provided, can write back\n")
	} else {
		fmt.Printf("NOTION_TOKEN not provided, read only\n")
	}
}

var (
	nProcessed            = 0
	nNotionPagesFromCache = 0
	nDownloadedPages      = 0
)

func eventObserver(ev interface{}) {
	switch v := ev.(type) {
	case *caching_downloader.EventError:
		log(v.Error)
	case *caching_downloader.EventDidDownload:
		nProcessed++
		nDownloadedPages++
		log("%03d '%s' : downloaded in %s\n", nProcessed, v.PageID, v.Duration)
	case *caching_downloader.EventDidReadFromCache:
		nProcessed++
		nNotionPagesFromCache++
		// TODO: only verbose
		//log("%03d '%s' : read from cache in %s\n", nProcessed, v.PageID, v.Duration)
	case *caching_downloader.EventGotVersions:
		log("downloaded info about %d versions in %s\n", v.Count, v.Duration)
	}
}

func downloadBook(c *notionapi.Client, book *Book) {
	log("Loading %s...\n", book.Title)
	nProcessed = 0
	nNotionPagesFromCache = 0
	nDownloadedPages = 0

	cacheDir := book.NotionCacheDir()
	dirCache, err := caching_downloader.NewDirectoryCache(cacheDir)
	must(err)
	d := caching_downloader.New(dirCache, c)
	d.EventObserver = eventObserver
	d.RedownloadNewerVersions = true
	d.NoReadCache = flgDisableNotionCache

	startPageID := book.NotionStartPageID
	pages, err := d.DownloadPagesRecursively(startPageID)
	must(err)
	for _, page := range pages {
		id := toNoDashID(page.ID)
		page := &Page{
			NotionPage: page,
		}
		book.idToPage[id] = page
	}

	log("Got %d pages for %s, downloaded: %d, from cache: %d\n", len(book.idToPage), book.Title, nDownloadedPages, nNotionPagesFromCache)
	bookFromPages(book)
}

func loadSOUserMappingsMust() {
	path := filepath.Join("stack-overflow-docs-dump", "users.json.gz")
	err := common.JSONDecodeGzipped(path, &soUserIDToNameMap)
	u.PanicIfErr(err)
}

func shouldCopyImage(path string) bool {
	return !strings.Contains(path, "@2x")
}

func copyCoversMust() {
	srcDir := "covers"
	dstDir := filepath.Join("www", "covers")
	copyFilesRecur(dstDir, srcDir, shouldCopyImage)
	dstDir = filepath.Join("www", "covers_small")
	srcDir = filepath.Join("covers", "covers_small")
	copyFilesRecur(dstDir, srcDir, shouldCopyImage)
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
	case "index.css":
		dstPtr = &pathIndexCSS
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

func genBooks(books []*Book) {
	timeStart := time.Now()
	clearSitemapURLS()
	copyCoversMust()

	copyToWwwAsSha1MaybeMust("main.css")
	copyToWwwAsSha1MaybeMust("index.css")
	copyToWwwAsSha1MaybeMust("app.js")
	copyToWwwAsSha1MaybeMust("favicon.ico")
	_ = genIndex(books, nil)
	_ = genIndexGrid(books, nil)
	gen404TopLevel()
	_ = genAbout(nil)
	_ = genFeedback(nil)

	if true {
		// parallel
		n := runtime.NumCPU()
		sem := make(chan bool, n)
		var wd sync.WaitGroup
		for _, book := range books {
			wd.Add(1)
			go func(b *Book) {
				sem <- true
				genBook(b)
				<-sem
				wd.Done()
			}(book)
		}
		wd.Wait()
	} else {
		for _, book := range books {
			genBook(book)
		}
	}
	writeSitemap()
	fmt.Printf("Finished generating all books in %s\n", time.Since(timeStart))
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
	doMinify = !flgPreviewStatic
}

func initBook(book *Book) {
	var err error

	createDirMust(book.OutputCacheDir())
	createDirMust(book.NotionCacheDir())

	if false {
		loadCache("cache/go/cache.txt")
		os.Exit(0)
	}

	book.idToPage = map[string]*Page{}
	book.cache = loadCache(book.cachePath())
	must(err)
}

func findBook(id string) *Book {
	for _, book := range allBooks {
		// fuzzy match - whatever hits
		parts := []string{book.Title, book.Dir, book.NotionStartPageID}
		for _, s := range parts {
			if strings.EqualFold(s, id) {
				return book
			}
		}
	}
	return nil
}

func adHoc() {
	if false {
		glotRunTestAndExit()
	}
	if false {
		glotGetSnippedIDTestAndExit()
	}

	// only needs to be run when we add new covers
	if false {
		genTwitterImagesAndExit()
	}
	if false {
		genSmallCoversAndExit()
	}
}

func isPreview() bool {
	return flgPreviewStatic || flgPreviewOnDemand
}

func main() {
	openLog()
	defer closeLog()

	parseFlags()

	adHoc()

	notionapi.LogFunc = log

	_ = os.RemoveAll("www")
	createDirMust(filepath.Join("www", "s"))
	createDirMust("log")

	timeStart := time.Now()

	initMinify()
	loadSOUserMappingsMust()

	if flgReportExternalLinks || flgReportStackOverflowLinks {
		reportExternalLinks()
		return
	}

	client := &notionapi.Client{
		AuthToken: notionAuthToken,
	}

	books := booksMain
	if flgAllBooks {
		books = allBooks
		log("Downloading all books\n")
	} else {
		if len(flag.Args()) > 0 {
			var newBooks []*Book
			for _, name := range flag.Args() {
				book := findBook(name)
				if book == nil {
					log("Didn't find book named '%s'\n", name)
					continue
				}
				newBooks = append(newBooks, book)
			}
			if len(newBooks) > 0 {
				books = newBooks
				log("Downloading %d books", len(books))
				for _, b := range books {
					log(" %s", b.Title)
				}
				log("\n")
			}
		}
	}
	for _, book := range books {
		initBook(book)
		downloadBook(client, book)
		loadSoContributorsMust(book)
	}
	log("Downloaded %d pages, %d from cache, in %s\n", nTotalDownloaded, nTotalFromCache, time.Since(timeStart))

	if flgPreviewOnDemand {
		log("Time: %s\n", time.Since(timeStart))
		startPreviewOnDemand(books)
		return
	}

	genStartTime := time.Now()
	genBooks(books)
	genNetlifyHeaders()
	genNetlifyRedirects(books)
	printAndClearErrors()
	log("Gen time: %s, total time: %s\n", time.Since(genStartTime), time.Since(timeStart))

	if flgPreviewStatic {
		startPreviewStatic()
	}
}
