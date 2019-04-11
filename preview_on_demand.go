package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	gPreviewBooks []*Book
)

func tryPrefixInDir(uri string, prefix string, dir string) string {
	if !strings.HasPrefix(uri, prefix) {
		return ""
	}
	uri = strings.TrimPrefix(uri, prefix)
	name := filepath.FromSlash(uri)
	path := filepath.Join(dir, name)
	if fileExists(path) {
		return path
	}
	fmt.Printf("tried path: '%s', name: '%s', uri: '%s'\n", path, name, uri)
	return ""
}

func getFileForURL(uri string) string {
	{
		path := tryPrefixInDir(uri, "/s/", tmplDir)
		if path != "" {
			return path
		}
	}
	{
		coversDir := filepath.Join("covers", "covers_small")
		path := tryPrefixInDir(uri, "/covers_small/", coversDir)
		if path != "" {
			return path
		}
	}
	return ""
}

// see if there's an exact match for this uri in tmpl folder
func serveFileFromTmpl(w http.ResponseWriter, r *http.Request) bool {
	uri := r.URL.Path
	path := getFileForURL(uri)
	if path == "" {
		return false
	}
	fmt.Printf("Served '%s' from file '%s'\n", r.URL.Path, path)
	http.ServeFile(w, r, path)
	return true
}

func writeHTMLHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
}

func handleIndexOnDemand(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("uri: %s\n", r.URL.Path)
	uri := r.URL.Path

	if uri == "/" {
		writeHTMLHeaders(w)
		err := genIndex(gPreviewBooks, w)
		logIfError(err)
		return
	}

	if uri == "/index-grid" {
		writeHTMLHeaders(w)
		err := genIndexGrid(gPreviewBooks, w)
		logIfError(err)
		return
	}

	if uri == "/about" {
		writeHTMLHeaders(w)
		err := genAbout(w)
		logIfError(err)
		return
	}

	if uri == "/feedback" {
		writeHTMLHeaders(w)
		err := genFeedback(w)
		logIfError(err)
		return
	}

	if serveFileFromTmpl(w, r) {
		return
	}
	serve404(w, r)
}

// https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
func makeHTTPServerOnDemand() *http.Server {
	mux := &http.ServeMux{}

	mux.HandleFunc("/", handleIndexOnDemand)

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second, // introduced in Go 1.8
		Handler:      mux,
	}
	return srv
}

func startPreviewOnDemand(books []*Book) {
	gPreviewBooks = books

	httpSrv := makeHTTPServerOnDemand()
	httpSrv.Addr = "127.0.0.1:8173"

	go func() {
		err := httpSrv.ListenAndServe()
		// mute error caused by Shutdown()
		if err == http.ErrServerClosed {
			err = nil
		}
		panicIfErr(err)
		fmt.Printf("HTTP server shutdown gracefully\n")
	}()
	fmt.Printf("Started listening on %s, %d books\n", httpSrv.Addr, len(books))
	openBrowser("http://" + httpSrv.Addr)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt /* SIGINT */, syscall.SIGTERM)
	sig := <-c
	fmt.Printf("Got signal %s\n", sig)
}
