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
func (c *Sha1ToGlotPlaygroundCache) GetPlaygroundID(d []byte, snippetName string, fileName string, lang string) (string, error) {
	sha1 := u.Sha1HexOfBytes(d)
	id, ok := c.sha1ToID[sha1]
	if ok {
		return id, nil
	}
	rsp, err := glotGetSnippedID(d, snippetName, fileName, lang)
	if err != nil {
		return "", err
	}
	id = rsp.ID
	s := fmt.Sprintf("%s %s\n", sha1, id)
	err = appendToFile(c.cachePath, s)
	if err != nil {
		return "", err
	}
	c.nUpdates++
	return id, nil
}

func getSha1ToGlotPlaygroundIDCached(b *Book, d []byte, snippetName string, fileName string, lang string) (string, error) {
	nUpdates := b.sha1ToGlotPlaygroundCache.nUpdates
	id, err := b.sha1ToGlotPlaygroundCache.GetPlaygroundID(d, snippetName, fileName, lang)
	if err != nil {
		return "", err
	}
	if nUpdates != b.sha1ToGlotPlaygroundCache.nUpdates {
		sha1 := u.Sha1HexOfBytes(d)
		fmt.Printf("getSha1ToGlotPlaygroundIDCached: %s => %s\n", sha1, id)
	}
	return id, nil
}
