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
	fmt.Printf("Loaded '%s' with %d entries\n", path, len(res.sha1ToID))
	return res
}

// GetPlaygroundID gets go playground id from content
func GetPlaygroundID(b *Book, d []byte, snippetName string, fileName string, lang string) (string, bool, error) {
	sha1 := u.Sha1HexOfBytes(d)
	id, ok := b.cache.sha1ToGlotID[sha1]
	if ok {
		return id, true, nil
	}
	rsp, err := glotGetSnippedID(d, snippetName, fileName, lang)
	if err != nil {
		return "", true, err
	}
	err = b.cache.addGlotSha1ToID(sha1, rsp.ID)
	if err != nil {
		return "", true, err
	}
	return id, false, nil
}

func getSha1ToGlotPlaygroundIDCached(b *Book, d []byte, snippetName string, fileName string, lang string) (string, error) {
	id, fromCache, err := GetPlaygroundID(b, d, snippetName, fileName, lang)
	if err != nil {
		return "", err
	}
	if !fromCache {
		sha1 := u.Sha1HexOfBytes(d)
		uri := "https://glot.io/snippets/" + id
		fmt.Printf("getSha1ToGlotPlaygroundIDCached: %s => %s, %s\n", sha1, id, uri)
	}
	return id, nil
}
