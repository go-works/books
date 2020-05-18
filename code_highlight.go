package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/kjk/u"
)

var (
	htmlFormatter  *html.Formatter
	highlightStyle *chroma.Style
)

// CodeBlockInfo represents info about code snippet
type CodeBlockInfo struct {
	Lang          string
	PlaygroundURI string
}

func init() {
	htmlFormatter = html.New(html.WithClasses(), html.TabWidth(2))
	u.PanicIf(htmlFormatter == nil, "couldn't create html formatter")
	styleName := "monokailight"
	highlightStyle = styles.Get(styleName)
	u.PanicIf(highlightStyle == nil, "didn't find style '%s'", styleName)
}

// gross hack: we need to change html generated by chroma
func fixupHTMLCodeBlock(htmlCode string, info *CodeBlockInfo) string {
	classLang := ""
	if info.Lang != "" {
		classLang = " lang-" + info.Lang
	}

	playgroundPart := ""
	if info.PlaygroundURI != "" {
		playgroundPart = fmt.Sprintf(`
<div class="code-box-playground">
	<a href="%s" target="_blank">try online</a>
</div>
`, info.PlaygroundURI)
	}

	html := fmt.Sprintf(`
<div class="code-box%s">
	<div>
	%s
	</div>
	<div class="code-box-nav">
		%s
	</div>
</div>`, classLang, htmlCode, playgroundPart)
	return html
}

// based on https://github.com/alecthomas/chroma/blob/master/quick/quick.go
func htmlHighlight2(w io.Writer, source, lang, defaultLang string) error {
	if lang == "" {
		lang = defaultLang
	}
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return htmlFormatter.Format(w, highlightStyle, it)
}

func htmlHighlight(w io.Writer, source, lang, defaultLang string) error {
	ch := make(chan error, 1)
	go func() {
		err := htmlHighlight2(w, source, lang, defaultLang)
		ch <- err
	}()
	reportOvertime := func() {
		ioutil.WriteFile("hili_hang.txt", []byte(source), 0644)
		logf("Too long processing lang: %s, defaultLang: %s, source:\n%s\n\n", lang, defaultLang, source)
		ch <- nil
	}
	timer := time.AfterFunc(time.Second*5, reportOvertime)
	defer timer.Stop()

	err := <-ch
	return err

}

func testHang() {
	//d, err := ioutil.ReadFile("hili_hang.txt")
	//must(err)
	// s := string(d)
	s := `// 64-bit floats have 53 digits of precision, including the whole-number-part.
double a =     0011111110111001100110011001100110011001100110011001100110011010; // imperfect representation of 0.1
double b =     0011111111001001100110011001100110011001100110011001100110011010; // imperfect representation of 0.2
double c =     0011111111010011001100110011001100110011001100110011001100110011; // imperfect representation of 0.3
double a + b = 0011111111010011001100110011001100110011001100110011001100110100; // Note that this is not quite equal to the "canonical" 0.3!a
`

	for i := 0; i < 1024*32; i++ {
		var buf bytes.Buffer
		htmlHighlight(&buf, s, "C++", "")
	}
}
