package main

/*
zeit doesn't work for me because it requires .html in the url
and not sure how I can do 404 redirect. The docs don't
mention a way to provide a route only when there's no
direct match.

maybe I can have a single index.js to respond to request.
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

const (
	useStatic = "@now/static"
)

type zeitBuild struct {
	Src string `json:"src"`
	Use string `json:"use"`
}

type zeitRoute struct {
	Src     string            `json:"src"`
	Dest    string            `json:"dest,omitempty"`
	Status  int               `json:"status,omitempty"`
	Methods []string          `json:"methods,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type zeitInfo struct {
	Name    string       `json:"name,omitempty"`
	Version int          `json:"version"`
	Builds  []*zeitBuild `json:"builds"`
	Routes  []*zeitRoute `json:"routes,omitempty"`
}

func genZeitInfo(dir string, books []*Book) {
	builds := []*zeitBuild{
		{"*.html", useStatic},
		{"*.js", useStatic},
		{"*.png", useStatic},
		{"*.txt", useStatic},
	}
	r1 := &zeitRoute{
		Src: "/s/.+",
		Headers: map[string]string{
			"cache-control": "max-age=31536000",
		},
	}
	routes := []*zeitRoute{
		r1,
	}

	for _, b := range books {
		src := fmt.Sprintf(`/essential/%s/.*`, b.Dir)
		dst := fmt.Sprintf(`/essential/%s/404.html`, b.Dir)
		r := &zeitRoute{
			Src:    src,
			Dest:   dst,
			Status: 404,
		}
		routes = append(routes, r)
	}

	zeit := zeitInfo{
		Name:    "essentialbooks",
		Version: 2,
		Builds:  builds,
		Routes:  routes,
	}
	path := filepath.Join(dir, "now.json")
	d, err := json.Marshal(zeit)
	must(err)
	err = ioutil.WriteFile(path, d, 0644)
	must(err)
}
