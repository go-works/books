package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kjk/u"
)

var (
	errTooManyRequests = errors.New("too many requests")
	httpClient         = http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

/*
ad-hoc code to create a list of stack overflow contributors for each book.
We only have user ids and we can convert them to names with:

$ http HEAD https://stackoverflow.com/users/1850609/
HTTP/1.1 301 Moved Permanently
Location: /users/1850609/acdcjunior
*/
func resolveUserName(id int, trace bool) (string, error) {
	uri := "https://stackoverflow.com/users/" + strconv.Itoa(id)
	res, err := httpClient.Head(uri)
	if err != nil {
		return "", err
	}
	// I assume 404 means that the user has been deleted
	if res.StatusCode == 404 {
		return "user_deleted", nil
	}
	if res.StatusCode == 429 {
		return "", errTooManyRequests
	}
	if res.StatusCode != 301 {
		fmt.Printf("%s\n", uri)
		fmt.Printf("%v\n", res.Header)
		return "", fmt.Errorf("status code: %d", res.StatusCode)
	}
	loc := res.Header.Get("Location")
	parts := strings.Split(loc, "/")
	lastIdx := len(parts) - 1
	name := parts[lastIdx]

	if trace {
		fmt.Printf("uri: '%s', loc: '%s', name: '%s'\n", uri, loc, name)
	}
	return name, nil
}

const userNamesPath = "users.json"

func loadUserNames() map[int]string {
	var res map[int]string
	f, err := os.Open(userNamesPath)
	if err != nil {
		return make(map[int]string)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(&res)
	if err != nil || res == nil {
		return make(map[int]string)
	}
	fmt.Printf("Loaded %d user names\n", len(res))
	return res
}

func saveUserNames(d map[int]string) {
	f, err := os.Create(userNamesPath)
	u.PanicIfErr(err)
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.Encode(d)
}

func resolveUserNames(userIds []int) map[int]string {
	res := loadUserNames()
	for _, userID := range userIds {
		if _, ok := res[userID]; ok {
			continue
		}
		name, err := resolveUserName(userID, true)
		if err != nil {
			saveUserNames(res)
			u.PanicIfErr(err)
			panic("saved user names")
		}
		res[userID] = name
		dur := time.Millisecond * 1000
		time.Sleep(dur)
	}
	saveUserNames(res)
	return res
}

func genContributorsAndExit() {
	timeStart := time.Now()
	fmt.Printf("Loading Stack Overflow data...")
	gDocTags := loadDocTagsMust()
	gContributors := loadContributorsMust()
	gTopics := loadTopicsMust()
	gExamples := loadExamplesMust()
	fmt.Printf(" took %s\n", time.Since(timeStart))

	// for fast search, build topic id => doc id mapping
	topicToBook := make(map[int]int)
	for _, topic := range gTopics {
		topicToBook[topic.Id] = topic.DocTagId
	}

	nMissingTopics := 0
	nMissingExamples := 0

	// for fast search, build example id => doc id mapping
	exampleToBook := make(map[int]int)
	for _, example := range gExamples {
		exampleID := example.Id
		topicID := example.DocTopicId
		bookID, ok := topicToBook[topicID]
		if !ok {
			//u.PanicIf(!ok, "missing topic id => book id in topicToBook, topicID: %d", topicID)
			nMissingTopics++
		}
		exampleToBook[exampleID] = bookID
	}

	bookToName := make(map[int]string)
	for _, book := range gDocTags {
		bookToName[book.Id] = book.Title
	}

	// maps book id to map of contributor id => count of contributions
	perBookContributors := make(map[int]map[int]int)

	contributors := make(map[int]int)
	var bookID int
	var ok bool
	for _, c := range gContributors {
		contributors[c.UserId]++
		topicID := c.DocTopicId
		if topicID != 0 {
			bookID, ok = topicToBook[topicID]
			if !ok {
				//u.PanicIf(!ok, "missing topic id = > book id in topicToBook, topicID: %d", topicID)
				nMissingTopics++
			}
		} else {
			exampleID := c.DocExampleId
			u.PanicIf(exampleID == 0, "both topicID and exampleID is 0")
			bookID, ok = exampleToBook[exampleID]
			if !ok {
				//u.PanicIf(!ok, "missing example id = > book id in exampleToBook, exampleID: %d", exampleID)
				nMissingExamples++
			}
		}
		bookContrib, ok := perBookContributors[bookID]
		if !ok {
			bookContrib = make(map[int]int)
			perBookContributors[bookID] = bookContrib
		}
		bookContrib[c.UserId]++
	}

	var contributorIds []int
	for uid := range contributors {
		contributorIds = append(contributorIds, uid)
	}
	fmt.Printf("Resolving user ids for %d users\n", len(contributorIds))
	userIDToName := resolveUserNames(contributorIds)

	fmt.Printf("Total contributions: %d\n", len(gContributors))
	fmt.Printf("Unique contributors: %d\n", len(contributors))
	fmt.Printf("Total books: %d\n", len(perBookContributors))
	fmt.Printf("Missing topics: %d\n", nMissingTopics)
	fmt.Printf("Missing examples: %d\n", nMissingExamples)

	fmt.Printf("All:\n")
	for cID, n := range contributors {
		uname := userIDToName[cID]
		fmt.Printf("%d, %d, %s\n", cID, n, uname)
	}
	for bookID, bookContributors := range perBookContributors {
		bookName := bookToName[bookID]
		fmt.Printf("%s:\n", bookName)
		for cID, n := range bookContributors {
			uname := userIDToName[cID]
			fmt.Printf("%d, %d, %s\n", cID, n, uname)
		}
	}
	os.Exit(0)
}
