package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kjk/notionapi"
)

var (
	useCacheForNotion = true
	// if true, we'll log
	logNotionRequests = true

	notionLogDir = "log"
)

// convert 2131b10c-ebf6-4938-a127-7089ff02dbe4 to 2131b10cebf64938a1277089ff02dbe4
func normalizeID(s string) string {
	return strings.Replace(s, "-", "", -1)
}

func openLogFileForPageID(pageID string) (io.WriteCloser, error) {
	if !logNotionRequests {
		return nil, nil
	}

	name := fmt.Sprintf("%s.go.log.txt", pageID)
	path := filepath.Join(notionLogDir, name)
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	return f, nil
}

func findSubPageIDs(blocks []*notionapi.Block) []string {
	pageIDs := map[string]struct{}{}
	seen := map[string]struct{}{}
	toVisit := blocks
	for len(toVisit) > 0 {
		block := toVisit[0]
		toVisit = toVisit[1:]
		id := normalizeID(block.ID)
		if block.Type == notionapi.BlockPage {
			pageIDs[id] = struct{}{}
			seen[id] = struct{}{}
		}
		for _, b := range block.Content {
			if b == nil {
				continue
			}
			id := normalizeID(block.ID)
			if _, ok := seen[id]; ok {
				continue
			}
			toVisit = append(toVisit, b)
		}
	}
	res := []string{}
	for id := range pageIDs {
		res = append(res, id)
	}
	sort.Strings(res)
	return res
}

func loadPageFromCache(b *Book, pageID string) *notionapi.Page {
	cachedPath := filepath.Join(b.NotionCacheDir(), pageID+".json")
	d, err := ioutil.ReadFile(cachedPath)
	if err != nil {
		return nil
	}

	var page notionapi.Page
	err = json.Unmarshal(d, &page)
	panicIfErr(err)
	return &page
}

// I got "connection reset by peer" error once so retry download 3 times, with a short sleep in-between
func downloadPageRetry(c *notionapi.Client, pageID string) (res *notionapi.Page, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\nRecovered with %v\n", r)
			fmt.Printf("PageID: %s\n\n", pageID)
			res = nil
			err = fmt.Errorf("crashed trying to downlaod page %s", pageID)
		}
	}()

	for i := 0; i < 3; i++ {
		if i > 0 {
			fmt.Printf("Download %s failed with '%s'\n", pageID, err)
			time.Sleep(3 * time.Second) // not sure if it matters
		}
		res, err = c.DownloadPage(pageID)
		if err == nil {
			return
		}
	}
	return
}

func downloadAndCachePage(c *notionapi.Client, b *Book, pageID string) (*notionapi.Page, error) {
	//fmt.Printf("downloading page with id %s\n", pageID)
	pageID = normalizeID(pageID)
	c.Logger, _ = openLogFileForPageID(pageID)
	if c.Logger != nil {
		defer func() {
			lf := c.Logger.(*os.File)
			lf.Close()
		}()
	}
	page, err := downloadPageRetry(c, pageID)
	if err != nil {
		return nil, err
	}
	d, err := json.MarshalIndent(page, "", "  ")
	if err != nil {
		// not a fatal error, just a warning
		fmt.Printf("json.Marshal() on pageID '%s' failed with %s\n", pageID, err)
		return page, nil
	}

	cachedPath := filepath.Join(b.NotionCacheDir(), pageID+".json")
	err = os.MkdirAll(filepath.Dir(cachedPath), 0755)
	panicIfErr(err)
	err = ioutil.WriteFile(cachedPath, d, 0644)
	panicIfErr(err)
	return page, nil
}

var (
	nNotionPagesFromCache int
)

func loadNotionPage(c *notionapi.Client, b *Book, pageID string, getFromCache bool, n int) (*notionapi.Page, error) {
	if b.isCachedPageNotOutdated[pageID] {
		page := b.cachedPagesFromDisk[pageID]
		fmt.Printf("Skipping %d %s %s (cached is current)\n", n, page.ID, page.Root.Title)
		return page, nil
	}

	if getFromCache {
		page := loadPageFromCache(b, pageID)
		if page != nil {
			nNotionPagesFromCache++
			//fmt.Printf("Got %d from cache %s %s\n", n, pageID, page.Root.Title)
			return page, nil
		}
	}

	page, err := downloadAndCachePage(c, b, pageID)
	if err == nil {
		fmt.Printf("Downloaded %d %s %s\n", n, page.ID, page.Root.Title)
	} else {
		return nil, err
	}

	//updated := updateFormatOrTitleIfNeeded(page)
	updated := false
	if !updated {
		return page, nil
	}

	time.Sleep(time.Millisecond * 100)
	page, err = downloadAndCachePage(c, b, pageID)
	if err == nil {
		fmt.Printf("Downloaded %d %s %s\n", n, page.ID, page.Root.Title)
	} else {
		return nil, err
	}

	return page, nil
}

func updateFormatIfNeeded(page *notionapi.Page) bool {
	// can't write back without a token
	if notionAuthToken == "" {
		return false
	}
	args := map[string]interface{}{}
	format := page.Root.FormatPage
	if format == nil || !format.PageSmallText {
		args["page_small_text"] = true
	}
	if format == nil || !format.PageFullWidth {
		args["page_full_width"] = true
	}
	if len(args) == 0 {
		return false
	}
	fmt.Printf("  updating format to %v\n", args)
	err := page.SetFormat(args)
	if err != nil {
		fmt.Printf("updateFormatIfNeeded: page.SetFormat() failed with '%s'\n", err)
	}
	return true
}

func updateTitleIfNeeded(page *notionapi.Page) bool {
	// can't write back without a token
	if notionAuthToken == "" {
		return false
	}
	newTitle := cleanTitle(page.Root.Title)
	if newTitle == page.Root.Title {
		return false
	}
	fmt.Printf("  updating title to '%s'\n", newTitle)
	err := page.SetTitle(newTitle)
	if err != nil {
		fmt.Printf("updateTitleIfNeeded: page.SetTitle() failed with '%s'\n", err)
	}
	return true
}

func updateFormatOrTitleIfNeeded(page *notionapi.Page) bool {
	updated1 := updateFormatIfNeeded(page)
	updated2 := updateTitleIfNeeded(page)
	return updated1 || updated2
}

func pageIDFromFileName(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return ""
	}
	id := parts[0]
	if len(id) == len("2b831bac5afc414493cff5e06e8e4460") {
		return id
	}
	return ""
}

func loadPagesFromDisk(b *Book) {
	dir := b.NotionCacheDir()
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("loadPagesFromDisk: os.ReadDir('%s') failed with '%s'\n", dir, err)
	}
	for _, f := range files {
		pageID := pageIDFromFileName(f.Name())
		if pageID == "" {
			continue
		}
		page := loadPageFromCache(b, pageID)
		b.cachedPagesFromDisk[pageID] = page
	}
	fmt.Printf("loadPagesFromDisk: loaded %d cached pages from %s\n", len(b.cachedPagesFromDisk), dir)
}

// convert:
// bb760e2dd6794b64b2a903005b21870a
// to:
// bb760e2d-d679-4b64-b2a9-03005b21870a
func toRequestID(id string) string {
	if len(id) == len("bb760e2d-d679-4b64-b2a9-03005b21870a") {
		return id
	}
	panicIf(len(id) != len("bb760e2dd6794b64b2a903005b21870a"), "invalid id '%s'", id)
	res := id[:8] + "-" + id[8:12] + "-" + id[12:16] + "-" + id[16:20] + "-" + id[20:]
	return res
}

func toRequestIDs(ids []string) []string {
	var res []string
	for _, id := range ids {
		res = append(res, toRequestID(id))
	}
	return res
}

func getVersionsForPages(c *notionapi.Client, ids []string) ([]int64, error) {
	// c.Logger = os.Stdout
	reqIDs := toRequestIDs(ids)
	recVals, err := c.GetRecordValues(reqIDs)
	if err != nil {
		return nil, err
	}
	results := recVals.Results
	if len(results) != len(ids) {
		return nil, fmt.Errorf("getVersionssForPages(): got %d results, expected %d", len(results), len(ids))
	}
	var versions []int64
	for i, res := range results {

		// this might happen e.g. when a page is not publicly visible
		if res.Value == nil {
			return nil, fmt.Errorf("Couldn't retrieve page with id %s", ids[i])
		}
		id, ok := notionapi.NormalizeID(res.Value.ID)
		panicIf(!ok)
		ver := res.Value.Version
		if ids[i] == id {
			panic("got result in the wrong order")
		}
		versions = append(versions, ver)
	}
	return versions, nil
}

func checkIfPagesAreOutdated(c *notionapi.Client, b *Book) {
	var ids []string
	for id := range b.cachedPagesFromDisk {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var versions []int64
	rest := ids
	maxPerCall := 128
	for len(rest) > 0 {
		n := len(rest)
		if n > maxPerCall {
			n = maxPerCall
		}
		tmpIDs := rest[:n]
		rest = rest[n:]
		fmt.Printf("getting versions for %d pages\n", len(tmpIDs))
		tmpVers, err := getVersionsForPages(c, tmpIDs)
		panicIfErr(err)
		versions = append(versions, tmpVers...)
	}
	panicIf(len(ids) != len(versions))
	nOutdated := 0
	for i, ver := range versions {
		id := ids[i]
		page := b.cachedPagesFromDisk[id]
		isOutdated := ver > page.Root.Version
		b.isCachedPageNotOutdated[id] = !isOutdated
		if isOutdated {
			nOutdated++
		}
	}
	fmt.Printf("checkIfPagesAreOutdated: %d pages, %d outdated\n", len(ids), nOutdated)
}

func loadNotionPages(c *notionapi.Client, b *Book, indexPageID string, idToPage map[string]*notionapi.Page, useCache bool) {
	loadPagesFromDisk(b)
	checkIfPagesAreOutdated(c, b)
	toVisit := []string{indexPageID}

	n := 1
	for len(toVisit) > 0 {
		pageID := normalizeID(toVisit[0])
		toVisit = toVisit[1:]

		if _, ok := idToPage[pageID]; ok {
			continue
		}

		page, err := loadNotionPage(c, b, pageID, useCache, n)
		panicIfErr(err)

		n++
		idToPage[pageID] = page

		subPages := findSubPageIDs(page.Root.Content)
		toVisit = append(toVisit, subPages...)
	}
}

func rmFile(path string) {
	err := os.Remove(path)
	if err != nil {
		fmt.Printf("os.Remove(%s) failed with %s\n", path, err)
	}
}

func rmCached(b *Book, pageID string) {
	id := normalizeID(pageID)
	rmFile(filepath.Join(notionLogDir, id+".go.log.txt"))
	rmFile(filepath.Join(b.NotionCacheDir(), id+".json"))
}

func createNotionLogDir() {
	if logNotionRequests {
		err := os.MkdirAll(notionLogDir, 0755)
		panicIfErr(err)
	}
}

func createNotionDirs() {
	createNotionLogDir()
}

func removeCachedNotion() {
	//err := os.RemoveAll(notionCacheDir)
	//panicIfErr(err)
	err := os.RemoveAll(notionLogDir)
	panicIfErr(err)
	createNotionDirs()
}
