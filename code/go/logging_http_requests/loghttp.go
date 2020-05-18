package main

import (
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kjk/dailyrotate"
	"github.com/kjk/siser"
)

var (
	httpLogDailyFile *dailyrotate.File
	httpLogSiser     *siser.Writer
)

// HTTPReqInfo describes info about HTTP request
type HTTPReqInfo struct {
	// GET etc.
	method  string
	url     string
	referer string
	ipaddr  string
	// response code, like 200, 404
	code int
	// number of bytes of the response sent
	size int64
	// how long did it take to
	duration  time.Duration
	userAgent string
}

func openHTTPLog() {
	dir := getHTTPLogDirMust()
	format := "2006-01-02.txt"
	path := filepath.Join(dir, format)
	var err error
	httpLogDailyFile, err = dailyrotate.NewFile(path, nil)
	must(err)
	httpLogSiser = siser.NewWriter(httpLogDailyFile)
}

func closeHTTPLog() {
	_ = httpLogDailyFile.Close()
	httpLogDailyFile = nil
	httpLogSiser = nil
}

var (
	muLogHTTP sync.Mutex
)

// we mostly care page views. to log less we skip logging
// of urls that don't provide useful information.
// hopefully we won't regret it
func skipHTTPRequestLogging(ri *HTTPReqInfo) bool {
	// we always want to know about failures and other
	// non-200 responses
	if ri.code != 200 {
		return false
	}

	// we want to know about slow requests.
	// 100 ms threshold is somewhat arbitrary
	if ri.duration > 100*time.Millisecond {
		return false
	}

	// this is linked from every page
	if ri.url == "/favicon.png" {
		return true
	}

	if ri.url == "/favicon.ico" {
		return true
	}

	if strings.HasSuffix(ri.url, ".css") {
		return true
	}
	return false
}

func logHTTPReq(ri *HTTPReqInfo) {
	if skipHTTPRequestLogging(ri) {
		return
	}

	var rec siser.Record
	rec.Name = "httplog"
	rec.Append("method", ri.method)
	rec.Append("uri", ri.url)
	if ri.referer != "" {
		rec.Append("referer", ri.referer)
	}
	rec.Append("ipaddr", ri.ipaddr)
	rec.Append("code", strconv.Itoa(ri.code))
	rec.Append("size", strconv.FormatInt(ri.size, 10))
	durMs := ri.duration / time.Millisecond
	rec.Append("duration", strconv.FormatInt(int64(durMs), 10))
	rec.Append("ua", ri.userAgent)

	muLogHTTP.Lock()
	defer muLogHTTP.Unlock()
	_, _ = httpLogSiser.WriteRecord(&rec)
}
