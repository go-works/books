package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/kjk/u"

	"github.com/essentialbooks/books/pkg/common"
)

// Sha1ToGoPlaygroundCache maintains sha1 of content to go playground id cache
type Sha1ToGoPlaygroundCache struct {
	cachePath string
	sha1ToID  map[string]string
	nUpdates  int
}

func readSha1ToGoPlaygroundCache(path string) *Sha1ToGoPlaygroundCache {
	res := &Sha1ToGoPlaygroundCache{
		cachePath: path,
		sha1ToID:  map[string]string{},
	}
	lines, err := common.ReadFileAsLines(path)
	if err != nil {
		if os.IsNotExist(err) {
			// early detection of "can't create a file" condition
			f, err := os.Create(path)
			panicIfErr(err)
			f.Close()
		}
	}
	for i, s := range lines {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		parts := strings.Split(s, " ")
		panicIf(len(parts) != 2, "unexpected line '%s'", lines[i])
		sha1 := parts[0]
		id := parts[1]
		res.sha1ToID[sha1] = id
	}
	fmt.Printf("Loaded '%s' with %d entries\n", path, len(res.sha1ToID))
	return res
}

// submit the data to Go playground and get share id
func getGoPlaygroundShareID(d []byte) (string, error) {
	uri := "https://play.golang.org/share"
	r := bytes.NewBuffer(d)
	resp, err := http.Post(uri, "text/plain", r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("http.Post returned error code '%s'", err)
	}
	d, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(d)), nil
}

func testGetGoPlaygroundShareIDAndExit() {
	path := "books/go/0230-mutex/rwlock.go"
	d, err := common.ReadFileNormalized(path)
	panicIfErr(err)
	shareID, err := getGoPlaygroundShareID(d)
	panicIfErr(err)
	fmt.Printf("share id: '%s'\n", shareID)
	os.Exit(0)
}

// GetGoPlaygroundID gets go playground id from content
func GetGoPlaygroundID(c *Cache, d []byte) (string, bool, error) {
	sha1 := u.Sha1HexOfBytes(d)
	id, ok := c.sha1ToGoPlayID[sha1]
	if ok {
		return id, true, nil
	}
	id, err := getGoPlaygroundShareID(d)
	if err != nil {
		return "", false, err
	}
	err = c.addGoPlaySha1ToID(sha1, id)
	if err != nil {
		return "", false, err
	}
	return id, false, nil
}

func getSha1ToGoPlaygroundIDCached(b *Book, d []byte) (string, error) {
	id, fromCache, err := GetGoPlaygroundID(b.cache, d)
	if err == nil && !fromCache {
		sha1 := u.Sha1HexOfBytes(d)
		fmt.Printf("getSha1ToGoPlaygroundIDCached: %s => %s\n", sha1, id)
	}
	return id, err
}
