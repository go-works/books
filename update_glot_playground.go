package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kjk/u"

	"github.com/essentialbooks/books/pkg/common"
)

// Sha1ToGlotPlaygroundCache maintains sha1 of content to go playground id cache
type Sha1ToGlotPlaygroundCache struct {
	cachePath string
	sha1ToID  map[string]string
	nUpdates  int
}

func readSha1ToGlotPlaygroundCache(path string) *Sha1ToGlotPlaygroundCache {
	res := &Sha1ToGlotPlaygroundCache{
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
	fmt.Printf("Loaded '%s' with %d glot sha1 => id entries\n", path, len(res.sha1ToID))
	return res
}

// GetGlotPlaygroundID gets go playground id from content
func GetGlotPlaygroundID(b *Book, d []byte, snippetName string, fileName string, lang string) (string, bool, error) {
	sha1 := u.Sha1HexOfBytes(d)
	id, ok := b.cache.sha1ToGlotID[sha1]
	if ok {
		//fmt.Printf("GetGlotPlaygroundID: got %s from cache\n", sha1)
		return id, true, nil
	}

	//fmt.Printf("GetGlotPlaygroundID: no %s in cache of %d entries\n", sha1, len(b.cache.sha1ToGlotID))
	rsp, err := glotGetSnippedID(d, snippetName, fileName, lang)
	if err != nil {
		return "", false, err
	}
	err = b.cache.addGlotSha1ToID(sha1, rsp.ID)
	if err != nil {
		return "", false, err
	}
	return id, false, nil
}

func getSha1ToGlotPlaygroundIDCached(b *Book, d []byte, snippetName string, fileName string, lang string) (string, error) {
	id, fromCache, err := GetGlotPlaygroundID(b, d, snippetName, fileName, lang)
	if err != nil {
		return "", err
	}
	if !fromCache {
		sha1 := u.Sha1HexOfBytes(d)
		uri := "https://glot.io/snippets/" + id
		fmt.Printf("getSha1ToGlotPlaygroundIDCached: generated glot id %s => %s, %s\n", sha1, id, uri)
	}
	return id, nil
}
