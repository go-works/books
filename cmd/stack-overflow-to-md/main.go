package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/essentialbooks/books/pkg/common"
	"github.com/essentialbooks/books/pkg/kvstore"
	"github.com/essentialbooks/books/pkg/stackoverflow"
	"github.com/kjk/u"
)

// DocTag is an alias
type DocTag = stackoverflow.DocTag

// Topic is an alias
type Topic = stackoverflow.Topic

// Example is an alias
type Example = stackoverflow.Example

// TopicHistory is an alias
type TopicHistory = stackoverflow.TopicHistory

// Contributor is an alias
type Contributor = stackoverflow.Contributor

var (
	flgPrintStats bool

	emptyExamplexs []*Example
	// if true, prints more information
	verbose = true

	booksToImport = common.BooksToProcess
)

func parseFlags() {
	flag.BoolVar(&flgPrintStats, "stats", false, "if true will show book stats")
	flag.Parse()
}

func getTopicsByDocID(docID int) map[int]bool {
	res := make(map[int]bool)
	topics := loadTopicsMust()
	for _, topic := range topics {
		if topic.DocTagId == docID {
			res[topic.Id] = true
		}
	}
	return res
}

func isEmptyString(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) == 0
}

func calcExampleCount(docTag *DocTag) {
	docID := docTag.Id
	topics := getTopicsByDocID(docID)
	n := 0
	examples := loadExamplesMust()
	for _, ex := range examples {
		if topics[ex.DocTopicId] {
			n++
		}
	}
	docTag.ExampleCount = n
}

func printDocTagsAndExit() {
	loadAll(false)
	docs := loadDocTagsMust()
	for i := range docs {
		docTag := &docs[i]
		calcExampleCount(docTag)
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].ExampleCount < docs[j].ExampleCount
	})
	for _, dc := range docs {
		fmt.Printf(`{ "%s", "", false, %d, %d },%s`, dc.Title, dc.ExampleCount, dc.TopicCount, "\n")
	}
	os.Exit(0)
}

var (
	docTagsCached        []DocTag
	topicsCached         []Topic
	topicHistoriesCached []TopicHistory
	contributorsCached   []*Contributor
	examplesCached       []*Example
)

func loadDocTagsMust() []DocTag {
	if docTagsCached == nil {
		var err error
		path := path.Join("stack-overflow-docs-dump", "doctags.json.gz")
		docTagsCached, err = stackoverflow.LoadDocTags(path)
		u.PanicIfErr(err)
	}
	return docTagsCached
}

func loadTopicHistoriesMust() []TopicHistory {
	if topicHistoriesCached == nil {
		var err error
		path := path.Join("stack-overflow-docs-dump", "topichistories.json.gz")
		topicHistoriesCached, err = stackoverflow.LoadTopicHistories(path)
		u.PanicIfErr(err)
	}
	return topicHistoriesCached
}

func loadContributorsMust() []*Contributor {
	if contributorsCached == nil {
		var err error
		path := path.Join("stack-overflow-docs-dump", "contributors.json.gz")
		contributorsCached, err = stackoverflow.LoadContibutors(path)
		u.PanicIfErr(err)
	}
	return contributorsCached
}

func loadTopicsMust() []Topic {
	if topicsCached == nil {
		var err error
		path := path.Join("stack-overflow-docs-dump", "topics.json.gz")
		topicsCached, err = stackoverflow.LoadTopics(path)
		u.PanicIfErr(err)
	}
	return topicsCached
}

func loadExamplesMust() []*Example {
	if examplesCached == nil {
		var err error
		path := path.Join("stack-overflow-docs-dump", "examples.json.gz")
		examplesCached, err = stackoverflow.LoadExamples(path)
		u.PanicIfErr(err)
	}
	return examplesCached
}

func findDocTagByTitleMust(docTags []DocTag, title string) DocTag {
	for _, dc := range docTags {
		if dc.Title == title {
			return dc
		}
	}
	logFatalf("Didn't find DocTag with title '%s'\n", title)
	return DocTag{}
}

func loadAll(silent bool) {
	timeStart := time.Now()
	if !silent {
		fmt.Printf("Loading Stack Overflow data...")
	}
	loadDocTagsMust()
	loadTopicsMust()
	loadExamplesMust()
	loadTopicHistoriesMust()
	loadContributorsMust()
	if !silent {
		fmt.Printf(" took %s\n", time.Since(timeStart))
	}
}

func getTopicsByDocTagID(docTagID int) []*Topic {
	gTopics := loadTopicsMust()
	var res []*Topic
	for i, topic := range gTopics {
		if topic.DocTagId == docTagID {
			res = append(res, &gTopics[i])
		}
	}
	return res
}

func getExampleByID(id int) *Example {
	gExamples := loadExamplesMust()
	for i, e := range gExamples {
		if e.Id == id {
			return gExamples[i]
		}
	}
	return nil
}

func getExamplesForTopic(docTagID int, docTopicID int) []*Example {
	gTopicHistories := loadTopicHistoriesMust()
	var res []*Example
	seenIds := make(map[int]bool)
	for _, th := range gTopicHistories {
		if th.DocTagId == docTagID && th.DocTopicId == docTopicID {
			id := th.DocExampleId
			if seenIds[id] {
				continue
			}
			seenIds[id] = true
			ex := getExampleByID(id)
			if ex == nil {
				//fmt.Printf("Didn't find example, docTagID: %d, docTopicID: %d\n", docTagID, docTopicID)
			} else {
				res = append(res, ex)
			}
		}
	}
	return res
}

func sortExamples(a []*Example) {
	sort.Slice(a, func(i, j int) bool {
		if a[i].IsPinned {
			return true
		}
		if a[j].IsPinned {
			return false
		}
		return a[i].Score > a[j].Score
	})
}

// sometime json representation of versions is empty array, we want to skip those
func shortenVersion(s string) string {
	if s == "[]" {
		return ""
	}
	return s
}

func writeIndexTxtMust(path string, topic *Topic) {
	s := kvstore.Serialize("Title", topic.Title)
	s += kvstore.Serialize("Id", strconv.Itoa(topic.Id))
	versions := shortenVersion(topic.VersionsJson)
	s += kvstore.SerializeLong("Versions", versions)
	if isEmptyString(versions) {
		s += kvstore.SerializeLong("VersionsHtml", topic.HelloWorldVersionsHtml)
	}

	s += kvstore.SerializeLong("Introduction", topic.IntroductionMarkdown)
	if isEmptyString(topic.IntroductionMarkdown) {
		s += kvstore.SerializeLong("IntroductionHtml", topic.IntroductionHtml)
	}

	s += kvstore.SerializeLong("Syntax", topic.SyntaxMarkdown)
	if isEmptyString(topic.SyntaxMarkdown) {
		s += kvstore.SerializeLong("SyntaxHtml", topic.SyntaxHtml)
	}

	s += kvstore.SerializeLong("Parameters", topic.ParametersMarkdown)
	if isEmptyString(topic.ParametersMarkdown) {
		s += kvstore.SerializeLong("ParametersHtml", topic.ParametersHtml)
	}

	s += kvstore.SerializeLong("Remarks", topic.RemarksMarkdown)
	if isEmptyString(topic.RemarksMarkdown) {
		s += kvstore.SerializeLong("RemarksHtml", topic.RemarksHtml)
	}

	createDirForFileMust(path)
	err := ioutil.WriteFile(path, []byte(s), 0644)
	u.PanicIfErr(err)
	if verbose {
		fmt.Printf("Wrote %s, %d bytes\n", path, len(s))
	}
}

func getIndexMd(topic *Topic) []byte {
	var s string

	if false {
		s += "---\n"
		s += kvstore.Serialize("Title", topic.Title)
		s += kvstore.Serialize("Id", strconv.Itoa(topic.Id))
		s += "---\n"
	}

	versions := shortenVersion(topic.VersionsJson)

	if !isEmptyString(versions) {
		s += "## Versions\n"
		s += versions
		s += "\n"
	} else if !isEmptyString(topic.HelloWorldVersionsHtml) {
		s += "## Versions HTML\n\n"
		s += topic.HelloWorldVersionsHtml
		s += "\n"
	}

	if !isEmptyString(topic.IntroductionMarkdown) {
		s += "## Introduction\n"
		s += topic.IntroductionMarkdown
		s += "\n"
	} else if !isEmptyString(topic.IntroductionHtml) {
		s += "## Introduction HTML\n\n"
		s += topic.IntroductionHtml
		s += "\n"
	}

	if !isEmptyString(topic.SyntaxMarkdown) {
		s += "## Syntax\n"
		s += topic.SyntaxMarkdown
		s += "\n"
	} else if !isEmptyString(topic.SyntaxHtml) {
		s += "## Syntax HTML\n"
		s += topic.SyntaxHtml
		s += "\n"
	}

	if !isEmptyString(topic.ParametersMarkdown) {
		s += "## Parameters\n"
		s += topic.ParametersMarkdown
		s += "\n"
	} else if !isEmptyString(topic.ParametersHtml) {
		s += "## Parameters HTML\n\n"
		s += topic.ParametersHtml
		s += "\n"
	}

	if !isEmptyString(topic.RemarksMarkdown) {
		s += "## Remarks\n"
		s += topic.RemarksMarkdown
		s += "\n"
	} else if !isEmptyString(topic.RemarksHtml) {
		s += "## Remarks HTML\n\n"
		s += topic.RemarksHtml
		s += "\n"
	}
	return []byte(s)
}

func writeIndexTxtMdMust(path string, topic *Topic, format bool) {
	d := getIndexMd(topic)
	createDirForFileMust(path)
	err := ioutil.WriteFile(path, d, 0644)
	u.PanicIfErr(err)
	if verbose {
		fmt.Printf("Wrote %s, %d bytes\n", path, len(d))
	}
	if !format {
		return
	}
	err = mdfmtFile(path)
	u.PanicIfErr(err)
}

func writeArticleMust(path string, example *Example) {
	s := kvstore.Serialize("Title", example.Title)
	s += kvstore.Serialize("Id", strconv.Itoa(example.Id))
	s += kvstore.Serialize("Score", strconv.Itoa(example.Score))
	s += kvstore.SerializeLong("Body", example.BodyMarkdown)
	if isEmptyString(example.BodyMarkdown) {
		s += kvstore.SerializeLong("BodyHtml", example.BodyHtml)
	}

	createDirForFileMust(path)
	err := ioutil.WriteFile(path, []byte(s), 0644)
	u.PanicIfErr(err)
	if verbose {
		fmt.Printf("Wrote %s, %d bytes\n", path, len(s))
	}
}

func writeArticleMdMust(path string, example *Example, format bool) {
	var s string
	isHTML := false
	if isEmptyString(example.BodyMarkdown) {
		path = strings.Replace(path, ".md", ".html", -1)
		s = example.BodyHtml
		isHTML = true
	} else {
		if false {
			s = "---\n"
			s += kvstore.Serialize("Title", example.Title)
			s += kvstore.Serialize("Id", strconv.Itoa(example.Id))
			s += kvstore.Serialize("Score", strconv.Itoa(example.Score))
			s += "---\n\n"
		}
		s += example.BodyMarkdown
	}
	d := []byte(s)

	createDirForFileMust(path)
	err := ioutil.WriteFile(path, d, 0644)
	u.PanicIfErr(err)
	if verbose {
		fmt.Printf("Wrote %s, %d bytes\n", path, len(d))
	}
	if isHTML || !format {
		return
	}
	err = mdfmtFile(path)
	u.PanicIfErr(err)
}

func printEmptyExamples() {
	for _, ex := range emptyExamplexs {
		fmt.Printf("empty example: %s, len(BodyHtml): %d\n", ex.Title, len(ex.BodyHtml))
	}
}

func getContributors(docID int) []int {
	gContributors := loadContributorsMust()
	topics := getTopicsByDocID(docID)
	contributors := make(map[int]bool)
	for _, c := range gContributors {
		topicID := c.DocTopicId
		if _, ok := topics[topicID]; ok {
			contributors[c.UserId] = true
		}
	}
	var res []int
	for id := range contributors {
		res = append(res, id)
	}
	return res
}

func genContributors(bookDstDir string, docID int) {
	contributors := getContributors(docID)
	var a []string
	for _, id := range contributors {
		a = append(a, strconv.Itoa(id))
	}
	s := strings.Join(a, "\n")
	path := filepath.Join(bookDstDir, "so_contributors.txt")
	createDirForFileMust(path)
	err := ioutil.WriteFile(path, []byte(s), 0644)
	u.PanicIfErr(err)
	//fmt.Printf("Wrote %s\n", path)
}

// A list of characters we consider separators in normal strings and replace with our canonical separator - rather than removing.
var (
	separators = regexp.MustCompile(`[&_=+:]`)

	dashes = regexp.MustCompile(`[\-]+`)
	spaces = regexp.MustCompile(`[\ ]+`)
)

func cleanString(s string, r *regexp.Regexp) string {
	// Remove any trailing space to avoid ending on -
	s = strings.Trim(s, " ")

	// Flatten accents first so that if we remove non-ascii we still get a legible name
	//s = Accents(s)

	// Replace certain joining characters with a dash
	s = separators.ReplaceAllString(s, " ")

	// Remove all other unrecognised characters - NB we do allow any printable characters
	s = r.ReplaceAllString(s, "")

	// Remove any multiple spaces caused by replacements above
	s = spaces.ReplaceAllString(s, " ")

	return s
}

var illegalName = regexp.MustCompile(`[^[:alnum:]-. ]`)

// limited sanitizing for my limited needs
// makes sure s can be used as a file name (not a path!)
func cleanFileName(s string) string {
	return cleanString(s, illegalName)
}

func writeBookAsMarkdown(docTag *DocTag, bookName string) {
	timeStart := time.Now()

	bookNameSafe := common.MakeURLSafe(bookName)
	bookTopDir := filepath.Join("books", bookNameSafe)
	bookTopDirFmt := filepath.Join("books", bookNameSafe+"_fmt")
	if u.PathExists(bookTopDir) {
		fmt.Printf("Book '%s' has already been imported.\nTo re-import, delete directory '%s'\n", bookName, bookTopDir)
		os.Exit(1)
	}
	if u.PathExists(bookTopDirFmt) {
		fmt.Printf("Book '%s' has already been imported.\nTo re-import, delete directory '%s'\n", bookName, bookTopDirFmt)
		os.Exit(1)
	}

	fmt.Printf("Importing a book %s\n", bookName)
	loadAll(false)

	//fmt.Printf("%s: docID: %d\n", title, docTag.Id)
	topics := getTopicsByDocTagID(docTag.Id)
	nChapters := len(topics)
	nArticles := 0
	chapter := 10
	for _, t := range topics {
		examples := getExamplesForTopic(docTag.Id, t.Id)
		sortExamples(examples)

		dirChapter := fmt.Sprintf("%04d-%s", chapter, common.MakeURLSafe(t.Title))

		dirPath := filepath.Join(bookTopDir, dirChapter)
		{
			chapterIndexPath := filepath.Join(dirPath, fmt.Sprintf("000 %s.md", cleanFileName(t.Title)))
			writeIndexTxtMdMust(chapterIndexPath, t, false)
		}

		dirPathFmt := filepath.Join(bookTopDirFmt, dirChapter)
		{
			chapterIndexPath := filepath.Join(dirPathFmt, fmt.Sprintf("000 %s.md", cleanFileName(t.Title)))
			writeIndexTxtMdMust(chapterIndexPath, t, true)
		}

		//fmt.Printf("%s\n", dirChapter)
		chapter += 10
		//fmt.Printf("%s, %d examples (%d), %s\n", t.Title, t.ExampleCount, len(examples), fileName)

		articleNo := 10
		for _, ex := range examples {
			if isEmptyString(ex.BodyMarkdown) && isEmptyString(ex.BodyHtml) {
				emptyExamplexs = append(emptyExamplexs, ex)
				continue
			}
			fileName := fmt.Sprintf("%03d %s.md", articleNo, cleanFileName(ex.Title))

			{
				path := filepath.Join(dirPath, fileName)
				writeArticleMdMust(path, ex, false)
			}

			{
				path := filepath.Join(dirPathFmt, fileName)
				writeArticleMdMust(path, ex, true)
			}

			//fmt.Printf("  %s %s '%s'\n", ex.Title, pinnedStr, fileName)
			//fmt.Printf("  %03d-%s\n", articleNo, fileName)
			//fmt.Printf("  %s\n", fileName)
			articleNo += 10
		}
		nArticles += len(examples)
	}
	genContributors(bookTopDir, docTag.Id)

	fmt.Printf("Imported %s (%d chapters, %d articles) in %s\n", bookName, nChapters, nArticles, time.Since(timeStart))
}

func getImportedBooks() []string {
	books, err := common.GetDirs("books")
	u.PanicIfErr(err)
	return books
}

func getAllBooks() []string {
	docTags := loadDocTagsMust()
	var books []string
	for _, doc := range docTags {
		book := doc.Title
		books = append(books, book)
	}
	return books
}

func printAllBookNames() {
	all := getAllBooks()
	s := strings.Join(all, ", ")
	fmt.Printf("All books: %s\n", s)
}

func printUsageAndExit() {
	fmt.Printf("Usage: import-stack-overflow book-to-import\n")
	imported := getImportedBooks()
	if len(imported) > 0 {
		s := strings.Join(imported, ", ")
		fmt.Printf("Already imported: %s\n", s)
	}
	printAllBookNames()
	os.Exit(1)
}

// some custom fixups for stack overflow book name => our book name
var bookNameFixups = [][]string{
	{"Intel x86 Assembly Language & Microarchitecture", "Intel x86 assembly"},
	{"tensorflow", "TensorFlow"},
	{"react-native", "React Native"},
	{"postgresql", "PostgreSQL"},
	{"batch-file", "Batch file"},
	{"excel-vba", "Excel VBA"},
	{"html5-canvas", "HTML Canvas"},
	{"algorithm", "Algorithms"},
	{"meteor", "Meteor"},
}

// convert a stack overflow book name to our book name
func fixupBookName(soName string) string {
	// "Ruby Language" => "Ruby" etc.
	if strings.HasSuffix(soName, "Language") {
		s := strings.TrimSuffix(soName, "Language")
		return strings.TrimSpace(s)
	}
	// manual overrides
	for _, fixup := range bookNameFixups {
		if soName == fixup[0] {
			return fixup[1]
		}
	}
	return soName
}

func findBookByDocID(id int) *DocTag {
	docTags := loadDocTagsMust()
	for i, doc := range docTags {
		if id == doc.Id {
			return &docTags[i]
		}
	}
	return nil
}

func findBookByName(bookName string) *DocTag {
	nameNoCase := strings.ToLower(bookName)
	docTags := loadDocTagsMust()
	for i, doc := range docTags {
		titleNoCase := strings.ToLower(doc.Title)
		if nameNoCase == titleNoCase {
			return &docTags[i]
		}
	}
	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type bookStat struct {
	n        int
	docID    int
	bookName string
}

// prints how many articles in each book, sorted by article count
func printStats() {
	// fmt.Printf("printStats: starting\n")
	loadAll(true)

	topicIDToExampleCount := map[int]int{}
	examples := loadExamplesMust()
	for _, e := range examples {
		topicIDToExampleCount[e.DocTopicId]++
	}
	// fmt.Printf("len(topicIDToExampleCount): %d\n", len(topicIDToExampleCount))

	docIDToArticleCount := map[int]int{}
	topics := loadTopicsMust()
	for _, topic := range topics {
		nExamples := topicIDToExampleCount[topic.Id] + 1
		docIDToArticleCount[topic.DocTagId] += nExamples
	}
	//fmt.Printf("len(docIDToArticleCount): %d\n", len(docIDToArticleCount))

	a := []*bookStat{}
	for docID, n := range docIDToArticleCount {
		s := &bookStat{
			n:     n,
			docID: docID,
		}
		doc := findBookByDocID(docID)
		s.bookName = doc.Title
		a = append(a, s)
	}
	sort.Slice(a, func(i, j int) bool {
		return a[i].n < a[j].n
	})
	//fmt.Printf("len(a): %d\n", len(a))
	fmt.Printf("count,name\n")
	for _, stat := range a {
		fmt.Printf("%d,%s\n", stat.n, stat.bookName)
	}
}

func main() {
	// for ad-hoc operations uncomment one of those
	// genContributorsAndExit()
	// printDocTagsAndExit()

	parseFlags()
	if flgPrintStats {
		printStats()
		return
	}

	args := os.Args[1:]
	if len(args) != 1 {
		printUsageAndExit()
	}
	timeStart := time.Now()
	fmt.Printf("Trying to import book %s\n", args[0])

	bookName := args[0]
	doc := findBookByName(bookName)
	if doc == nil {
		printAllBookNames()
		fmt.Printf("\nDidn't find a book '%s'.\nSee above for list of available books\n", bookName)
		os.Exit(1)
	}

	writeBookAsMarkdown(doc, bookName)

	fmt.Printf("Took %s\n", time.Since(timeStart))
	//printEmptyExamples()
}
