package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kjk/siser"
)

// TODO: write me

type Cache struct {
	path string

	sha1ToGlotID   map[string]string
	sha1ToGoPlayID map[string]string
}

func (c *Cache) addGoPlaySha1ToID(sha1 string, id string) error {
	fmt.Printf("addGoPlaySha1ToID: %s => %s\n", sha1, id)
	// TODO: maybe silently skip?
	v, ok := c.sha1ToGoPlayID[sha1]
	panicIf(ok, "record already exists for sha1 '%s', value: '%s', new value: '%s'", sha1, v, id)
	c.sha1ToGoPlayID[sha1] = id
	r := siser.Record{
		Keys:   []string{"sha1", "id"},
		Values: []string{sha1, id},
		Name:   "goplayid",
	}
	f := openForAppend(c.path)
	defer f.Close()
	w := siser.NewWriter(f)
	w.Format = siser.FormatSizePrefix
	_, err := w.WriteRecord(&r)
	return err
}

func (c *Cache) addGlotSha1ToID(sha1 string, id string) error {
	fmt.Printf("addGlotSha1ToID: %s => %s\n", sha1, id)
	// TODO: maybe silently skip?
	v, ok := c.sha1ToGlotID[sha1]
	panicIf(ok, "record already exists for sha1 '%s', value: '%s', new value: '%s'", sha1, v, id)
	c.sha1ToGlotID[sha1] = id
	r := siser.Record{
		Keys:   []string{"sha1", "id"},
		Values: []string{sha1, id},
		Name:   "glotsha1",
	}
	f := openForAppend(c.path)
	defer f.Close()
	w := siser.NewWriter(f)
	w.Format = siser.FormatSizePrefix
	_, err := w.WriteRecord(&r)
	return err
}

func loadCache(path string) *Cache {
	fmt.Printf("loadCache: %s\n", path)
	dir := filepath.Dir(path)
	// the directory must exist
	_, err := os.Stat(dir)
	must(err)

	cache := &Cache{
		path:           path,
		sha1ToGlotID:   map[string]string{},
		sha1ToGoPlayID: map[string]string{},
	}

	f, err := os.Open(path)
	if err != nil {
		// it's ok if file doesn't exist
		fmt.Printf("Cache file %s doesn't exist\n", path)
		return cache
	}
	defer f.Close()

	r := siser.NewReader(f)
	r.Format = siser.FormatSizePrefix
	for r.ReadNext() {
		_, rec := r.Record()
		if rec.Name == "glotsha1" {
			sha1, ok := rec.Get("sha1")
			panicIf(!ok, "didn't find 'sha1' key in record named '%s'", rec.Name)
			id, ok := rec.Get("id")
			panicIf(!ok, "didn't find 'id' key in record named '%s'", rec.Name)
			cache.sha1ToGlotID[sha1] = id
			//fmt.Printf("glot %s => %s\n", sha1, id)
		} else if rec.Name == "goplayid" {
			sha1, ok := rec.Get("sha1")
			panicIf(!ok, "didn't find 'sha1' key in record named '%s'", rec.Name)
			id, ok := rec.Get("id")
			panicIf(!ok, "didn't find 'id' key in record named '%s'", rec.Name)
			cache.sha1ToGoPlayID[sha1] = id
			//fmt.Printf("goplay %s => %s\n", sha1, id)
		} else {
			panic(fmt.Errorf("unknown record: '%s'", rec.Name))
		}
	}
	must(r.Err())
	fmt.Printf("loadCache: len(cache.sha1ToGlotID): %d, len(cache.sha1ToGoPlayID): %d\n", len(cache.sha1ToGlotID), len(cache.sha1ToGoPlayID))
	return cache
}
