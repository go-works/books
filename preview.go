package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/kjk/u"
)

func fileForURI(uri string) string {
	path := filepath.Join("www", uri)
	if u.FileExists(path) {
		return path
	}
	path = path + ".html"
	if u.FileExists(path) {
		return path
	}
	return ""
}
func serve404(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	path := filepath.Join("www", "404.html")

	parts := strings.Split(uri[1:], "/")
	if len(parts) > 2 && parts[0] == "essential" {
		bookName := parts[1]
		maybePath := filepath.Join("www", "essential", bookName, "404.html")
		if u.FileExists(maybePath) {
			fmt.Printf("'%s' exists\n", maybePath)
			path = maybePath
		} else {
			fmt.Printf("'%s' doesn't exist\n", maybePath)
		}
	}
	fmt.Printf("Serving 404 from '%s' for '%s'\n", path, uri)
	d, err := ioutil.ReadFile(path)
	if err != nil {
		d = []byte(fmt.Sprintf("URL '%s' not found!", uri))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)
	w.Write(d)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	if uri == "/" {
		uri = "/index.html"
	}
	uriLocal := filepath.FromSlash(uri)
	if !strings.HasSuffix(uri, ".png") {
		fmt.Printf("uri: '%s'\n", uri)
	}
	path := fileForURI(uriLocal)
	if path == "" {
		serve404(w, r)
		return
	}
	http.ServeFile(w, r, path)
}

// https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
func makeHTTPServer() *http.Server {
	mux := &http.ServeMux{}

	mux.HandleFunc("/", handleIndex)

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second, // introduced in Go 1.8
		Handler:      mux,
	}
	return srv
}

func startPreviewStatic() {
	fmt.Printf("startPreviewStatic()\n")
	httpSrv := makeHTTPServer()
	httpSrv.Addr = "127.0.0.1:8173"

	go func() {
		err := httpSrv.ListenAndServe()
		// mute error caused by Shutdown()
		if err == http.ErrServerClosed {
			err = nil
		}
		u.Must(err)
		fmt.Printf("HTTP server shutdown gracefully\n")
	}()
	fmt.Printf("Started listening on %s\n", httpSrv.Addr)
	u.OpenBrowser("http://" + httpSrv.Addr)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt /* SIGINT */, syscall.SIGTERM)
	sig := <-c
	fmt.Printf("Got signal %s\n", sig)
}
