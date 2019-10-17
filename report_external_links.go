package main

import (
	"sort"
	"strings"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/caching_downloader"
)

func isNotionURL(uri string) bool {
	return strings.Contains(uri, "notion.so/")
}

func isStackOverflowURL(uri string) bool {
	return strings.Contains(uri, "stackoverflow.com/") && !strings.Contains(uri, "/users/")
}

func reportExternalLinksInPage(page *notionapi.Page) error {
	links := map[string]struct{}{}
	rememberLink := func(uri string) {
		if isNotionURL(uri) {
			return
		}
		if flgReportStackOverflowLinks {
			if isStackOverflowURL(uri) {
				links[uri] = struct{}{}
			}
			return
		}
		links[uri] = struct{}{}
	}
	findLinks := func(b *notionapi.Block) {
		spans := b.GetTitle()
		for _, ts := range spans {
			for _, attr := range ts.Attrs {
				attrType := notionapi.AttrGetType(attr)
				if attrType == notionapi.AttrLink {
					uri := notionapi.AttrGetLink(attr)
					rememberLink(uri)
				}
			}
		}
	}
	page.ForEachBlock(findLinks)
	if len(links) == 0 {
		return nil
	}
	id := toNoDashID(page.ID)
	logf("  page https://www.notion.so/%s has %d links\n", id, len(links))
	var a []string
	for uri := range links {
		a = append(a, uri)
	}
	sort.Strings(a)
	for _, uri := range a {
		logf("    %s\n", uri)
	}
	return nil
}

func reportExternalLinksInBook(book *Book) {
	cacheDir := book.NotionCacheDir()
	dirCache, err := caching_downloader.NewDirectoryCache(cacheDir)
	must(err)
	client := &notionapi.Client{
		AuthToken: notionAuthToken,
	}
	d := caching_downloader.New(dirCache, client)
	d.EventObserver = eventObserver
	d.RedownloadNewerVersions = false

	startPageID := book.NotionStartPageID
	nProcessed = 0
	nNotionPagesFromCache = 0
	nDownloadedPages = 0
	pages, err := d.DownloadPagesRecursively(startPageID, reportExternalLinksInPage)
	must(err)
	nPages := len(pages)
	log("Book %s, %d pages, downloaded: %d, from cache: %d\n", book.Title, nPages, nDownloadedPages, nNotionPagesFromCache)
}

func reportExternalLinks() {
	log("starting reportExternalLinks()\n")
	for _, b := range booksMain {
		reportExternalLinksInBook(b)
	}
}
