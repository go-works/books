package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	// https://www.netlify.com/docs/headers-and-basic-auth/#custom-headers
	netlifyHeaders = `
# long-lived caching
/s/*
  Cache-Control: max-age=31536000
/*
  X-Content-Type-Options: nosniff
  X-Frame-Options: DENY
  X-XSS-Protection: 1; mode=block
`
)

func genNetlifyHeaders() {
	path := filepath.Join("www", "_headers")
	err := ioutil.WriteFile(path, []byte(netlifyHeaders), 0644)
	panicIfErr(err)
}

// TODO: this should be in 404.html for each book
func genNetlifyRedirectsForBook(b *Book) []string {
	var res []string

	pages := b.GetAllPages()
	for _, page := range pages {
		id := page.NotionID
		uri := page.URLLastPath()
		s := fmt.Sprintf(`/essential/%s/%s* /essential/%s/%s 302`, b.Dir, id, b.Dir, uri)
		res = append(res, s)
	}

	if b.Dir == "go" {
		// only for Go book, add redirect from old ids to new ones
		for _, page := range pages {
			id := page.getID()
			if id == "" {
				continue
			}
			uri := page.URLLastPath()
			s := fmt.Sprintf(`/essential/%s/%s-* /essential/%s/%s 302`, b.Dir, id, b.Dir, uri)
			res = append(res, s)
		}
	}

	// catch-all redirect for all other missing pages
	s := fmt.Sprintf(`/essential/%s/* /essential/%s/404.html 404`, b.Dir, b.Dir)
	res = append(res, s)
	res = append(res, "")
	return nil
}

func genNetlifyRedirects(books []*Book) {
	var a []string
	for _, b := range books {
		ab := genNetlifyRedirectsForBook(b)
		a = append(a, ab...)
	}
	s := strings.Join(a, "\n")
	path := filepath.Join("www", "_redirects")
	err := ioutil.WriteFile(path, []byte(s), 0644)
	panicIfErr(err)
}
