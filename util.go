package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kjk/u"
)

func must(err error, args ...interface{}) {
	if err == nil {
		return
	}
	if len(args) == 0 {
		panic(err)
	}
	s := args[0].(string)
	if len(args) > 1 {
		args = args[1:]
		s = fmt.Sprintf(s, args)
	}
	panic(s + " err: " + err.Error())
}

func panicIfErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func panicMsg(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	fmt.Printf("%s\n", s)
	panic(s)
}

// FmtArgs formats args as a string. First argument should be format string
// and the rest are arguments to the format
func FmtArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}
	format := args[0].(string)
	if len(args) == 1 {
		return format
	}
	return fmt.Sprintf(format, args[1:]...)
}

func panicWithMsg(defaultMsg string, args ...interface{}) {
	s := FmtArgs(args...)
	if s == "" {
		s = defaultMsg
	}
	fmt.Printf("%s\n", s)
	panic(s)
}

func panicIf(cond bool, args ...interface{}) {
	if !cond {
		return
	}
	panicWithMsg("PanicIf: condition failed", args...)
}

func logIfError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

// whitelisted characters valid in url
func validateRune(c rune) byte {
	if c >= 'a' && c <= 'z' {
		return byte(c)
	}
	if c >= '0' && c <= '9' {
		return byte(c)
	}
	if c == '-' || c == '_' || c == '.' || c == ' ' {
		return '-'
	}
	return 0
}

func charCanRepeat(c byte) bool {
	if c >= 'a' && c <= 'z' {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	return false
}

// urlify generates safe url from tile by removing hazardous characters
func urlify(title string) string {
	s := strings.TrimSpace(title)
	s = strings.ToLower(s)
	var res []byte
	for _, r := range s {
		c := validateRune(r)
		if c == 0 {
			continue
		}
		// eliminute duplicate consequitive characters
		var prev byte
		if len(res) > 0 {
			prev = res[len(res)-1]
		}
		if c == prev && !charCanRepeat(c) {
			continue
		}
		res = append(res, c)
	}
	s = string(res)
	if len(s) > 128 {
		s = s[:128]
	}
	s = strings.TrimLeft(s, "-")
	s = strings.TrimRight(s, "-")
	return s
}

var (
	softErrorMode bool
	delayedErrors []string

	totalHTMLBytes         int
	totalHTMLBytesMinified int
)

func maybePanicIfErr(err error) {
	if err == nil {
		return
	}
	if !softErrorMode {
		panicIfErr(err)
	}
	delayedErrors = append(delayedErrors, err.Error())
}

func clearErrors() {
	delayedErrors = nil
	totalHTMLBytes = 0
	totalHTMLBytesMinified = 0
}

func printAndClearErrors() {
	fmt.Printf("HTML: optimized %d => %d (saved %d bytes)\n", totalHTMLBytes, totalHTMLBytesMinified, totalHTMLBytes-totalHTMLBytesMinified)
	if len(delayedErrors) == 0 {
		return
	}
	errStr := strings.Join(delayedErrors, "\n")
	fmt.Printf("\n%d errors:\n%s\n\n", len(delayedErrors), errStr)
	clearErrors()
}

func createDirForFileMaybeMust(path string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	maybePanicIfErr(err)
}

func copyFileMaybeMust(dst, src string) error {
	createDirForFileMaybeMust(dst)
	err := copyFile(dst, src)
	maybePanicIfErr(err)
	return err
}

// "foo.js" => "foo-${sha1}.js"
func nameToSha1Name(name, sha1Hex string) string {
	ext := filepath.Ext(name)
	n := len(name)
	s := name[:n-len(ext)]
	return s + "-" + sha1Hex[:8] + ext
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		logFatal("%s", err)
	}
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return st.Mode().IsRegular()
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDirectory(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func createDirMust(dir string) {
	err := os.MkdirAll(dir, 0755)
	panicIfErr(err)
}

func copyFile(dst, src string) error {
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()
	fout, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fout.Close()
	_, err = io.Copy(fout, fin)
	return err
}

func copyFileMust(dst, src string) {
	err := copyFile(dst, src)
	panicIfErr(err)
}

func copyFilesRecur(dstDir, srcDir string, shouldCopyFunc func(path string) bool) {
	createDirMust(dstDir)
	fileInfos, err := ioutil.ReadDir(srcDir)
	u.PanicIfErr(err)
	for _, fi := range fileInfos {
		name := fi.Name()
		if fi.IsDir() {
			dst := filepath.Join(dstDir, name)
			src := filepath.Join(srcDir, name)
			copyFilesRecur(dst, src, shouldCopyFunc)
			continue
		}

		src := filepath.Join(srcDir, name)
		dst := filepath.Join(dstDir, name)
		shouldCopy := true
		if shouldCopyFunc != nil {
			shouldCopy = shouldCopyFunc(src)
		}
		if !shouldCopy {
			continue
		}
		if pathExists(dst) {
			continue
		}
		copyFileMust(dst, src)
	}
}

func getDirsRecur(dir string) ([]string, error) {
	toVisit := []string{dir}
	idx := 0
	for idx < len(toVisit) {
		dir = toVisit[idx]
		idx++
		fileInfos, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, fi := range fileInfos {
			if !fi.IsDir() {
				continue
			}
			path := filepath.Join(dir, fi.Name())
			toVisit = append(toVisit, path)
		}
	}
	return toVisit, nil
}

// "foo" + "bar" = "foo/bar", only one "/"
func urlJoin(s1, s2 string) string {
	if strings.HasSuffix(s1, "/") {
		if strings.HasPrefix(s2, "/") {
			return s1 + s2[1:]
		}
		return s1 + s2
	}

	if strings.HasPrefix(s2, "/") {
		return s1 + s2
	}
	return s1 + "/" + s2
}

// removes empty lines from the beginning and end of the array
func trimEmptyLines(lines []string) []string {
	for len(lines) > 0 && len(lines[0]) == 0 {
		lines = lines[1:]
	}

	for len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	n := len(lines)
	res := make([]string, 0, n)
	prevWasEmpty := false
	for i := 0; i < n; i++ {
		l := lines[i]
		shouldAppend := l != "" || !prevWasEmpty
		prevWasEmpty = l == ""
		if shouldAppend {
			res = append(res, l)
		}
	}
	return res
}

func countStartChars(s string, c byte) int {
	for i := range s {
		if s[i] != c {
			return i
		}
	}
	return len(s)
}

// remove longest common space/tab prefix on non-empty lines
func shiftLines(lines []string) {
	maxTabPrefix := 1024
	maxSpacePrefix := 1024
	// first determine how much we can remove
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		n := countStartChars(line, ' ')
		if n > 0 {
			if n < maxSpacePrefix {
				maxSpacePrefix = n
			}
			continue
		}
		n = countStartChars(line, '\t')
		if n > 0 {
			if n < maxTabPrefix {
				maxTabPrefix = n
			}
			continue
		}
		// if doesn't start with space or tab, early abort
		return
	}
	if maxSpacePrefix == 1024 && maxTabPrefix == 1024 {
		return
	}

	toRemove := maxSpacePrefix
	if maxTabPrefix != 1024 {
		toRemove = maxTabPrefix
	}
	if toRemove == 0 {
		return
	}

	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		lines[i] = line[toRemove:]
	}
}

// replace potentially windows paths \foo\bar into unix paths /foo/bar
func toUnixPath(s string) string {
	return strings.Replace(s, `\`, "/", -1)
}

func dataToLines(d []byte) []string {
	s := string(d)
	return strings.Split(s, "\n")
}

func reverseStringSlice(a []string) {
	n := len(a) / 2
	for i := 0; i < n; i++ {
		a[i], a[n-i] = a[n-i], a[i]
	}
}

// turn "010 Defining a SetterGetter" to "Defining a SetterGetter"
func cleanTitle(s string) string {
	idx := strings.Index(s, " ")
	if idx == -1 {
		return s
	}
	if idx > 6 {
		return s
	}
	maybeNum := s[:idx]
	rest := s[idx+1:]
	_, err := strconv.Atoi(maybeNum)
	if err != nil {
		return s
	}
	return strings.TrimSpace(rest)
}

func openForAppend(path string) *os.File {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	must(err)
	return f
}

// appends a line to a file
func appendToFile(path string, s string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString(s)
	if err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func httpGet(uri string) ([]byte, error) {
	hc := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := hc.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		d, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Request was '%s' (%d) and not OK (200). Body:\n%s\nurl: %s", resp.Status, resp.StatusCode, string(d), uri)
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func unzipFileAsData(f *zip.File) ([]byte, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// fileClose() is like .Close() but ignores error,
// for use in defer
func fileClose(c io.Closer) {
	_ = c.Close()
}
