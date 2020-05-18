package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

var (
	flgHTTPAddr string
	flgDataDir  string

	dataDirCached = ""
)

func parseFlags() {
	flag.StringVar(&flgHTTPAddr, "http-addr", ":6789", "address on which to listen")
	flag.StringVar(&flgDataDir, "data-dir", "", "data directory")
	flag.Parse()
	if flgDataDir != "" {
		dataDirCached = flgDataDir
	}
	// for error early if data dir cannot be created
	dir := getDataDirMust()
	logf("Data dir: %s\n", dir)
}

func getDataDirMust() string {
	if dataDirCached != "" {
		return dataDirCached
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		// assume we're running in dev
		dir, err := os.UserHomeDir()
		must(err)
		dir = filepath.Join(dir, "data", "myservice")
		err = os.MkdirAll(dir, 0755)
		must(err)
		dataDirCached = dir
		return dataDirCached
	}

	// assume running in production, on linux server
	dir := "/data/myservice"
	dataDirCached = makeDirMust(dir)
	return dataDirCached
}

func getHTTPLogDirMust() string {
	dir := filepath.Join(getDataDirMust(), "log_http")
	return makeDirMust(dir)
}

// implement io.Signal interface
type dummySignal struct {
}

func (s *dummySignal) String() string {
	return "dummy signal"
}
func (s *dummySignal) Signal() {
	// no-op
}

func makeTestRequest(uri string) {
	_, err := http.Get(uri)
	must(err)
	fmt.Printf("Made HTTP GET request '%s'\n", uri)
}

// makes http requests to trigger logging
func makeTestHTTPRequestsAndSignal(ch chan os.Signal) {
	baseURL := "http://127.0.0.1" + flgHTTPAddr
	makeTestRequest(baseURL + "/index.html")
	makeTestRequest(baseURL + "/hello")

	// signal the channel to end the app
	ch <- &dummySignal{}
}

func main() {
	parseFlags()

	openHTTPLog()

	logf("Starting on addr: %v\n", flgHTTPAddr)

	httpSrv := makeHTTPServer()
	httpSrv.Addr = flgHTTPAddr

	// start http server
	chServerClosed := make(chan bool, 1)
	go func() {
		err := httpSrv.ListenAndServe()
		// mute error caused by Shutdown()
		if err == http.ErrServerClosed {
			err = nil
		}
		must(err)
		logf("HTTP server shutdown gracefully\n")
		chServerClosed <- true
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt /* SIGINT */, syscall.SIGTERM)
	// give the server time to start up
	// don't do it in production:
	time.Sleep(time.Second * 2)

	go makeTestHTTPRequestsAndSignal(c)

	sig := <-c
	logf("Got signal '%s'\n", sig)

	// cleanly shut down the server
	if httpSrv != nil {
		// Shutdown() needs a non-nil context
		_ = httpSrv.Shutdown(context.Background())
		select {
		case <-chServerClosed:
			// do nothing
		case <-time.After(time.Second * 5):
			// timeout
		}
	}
	logPath := httpLogDailyFile.Path()
	closeHTTPLog()

	// print http log to stdout
	d, err := ioutil.ReadFile(logPath)
	must(err)
	fmt.Printf("\nContent of the log file '%s':\n\n%s\n", logPath, string(d))
}
