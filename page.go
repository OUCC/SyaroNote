package main

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/util"
	"github.com/OUCC/syaro/wikiio"

	"github.com/russross/blackfriday"

	"bufio"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	PAGE_VIEW  = "page.html"
	SIDEBAR_MD = "_Sidebar.md"
)

var (
	ErrIsNotMarkdown = errors.New("requested file is not a markdown")
)

type Page struct {
	*wikiio.WikiFile
	markdownHTML template.HTML
}

// LoadPage returns new Page.
func LoadPage(wpath string) (*Page, error) {
	LoggerV.Printf("main.LoadPage(%s)\n", wpath)

	wfile, err := wikiio.Load(wpath)
	if err != nil {
		LoggerV.Println("Error: main.LoadPage:", err)
		return nil, err
	}

	// check if file isn't markdown
	if !(wfile.IsDir() || wfile.IsMarkdown()) {
		LoggerV.Println("Error: main.LoadPage: file isn't markdown")
		return nil, ErrIsNotMarkdown
	}

	LoggerV.Println("main.LoadPage: ok")
	return &Page{wfile, ""}, nil
}

// WikiPathList returns slice of each pages in wikipath
// (slice dosen't include urlPrefix)
func (page *Page) WikiPathList() []*Page {
	s := strings.Split(util.RemoveExt(page.WikiPath()), "/")
	if s[0] == "" {
		s = s[1:]
	}

	ret := make([]*Page, len(s))
	for i := 0; i < len(ret); i++ {
		ret[i], _ = LoadPage(filepath.Join(setting.WikiRoot, strings.Join(s[:i+1], "/")))
	}
	return ret
}

// Title returns title of page.
func (page *Page) Title() string {
	reader := strings.NewReader(string(page.MarkdownHTML()))
	scanner := bufio.NewScanner(reader)

	re := regexp.MustCompile("<h1>([^<]+)</h1>")
	for scanner.Scan() {
		submatch := re.FindStringSubmatch(scanner.Text())
		if len(submatch) != 0 {
			return submatch[1]
		}
	}
	return ""
}

// row returns row file data.
func (page *Page) raw() []byte {
	if page.IsDir() {
		LoggerV.Println("main.Page.raw: requested page is dir, use main file of dir")

		// wiki root. use Home
		var file *wikiio.WikiFile
		var err error
		if page.WikiPath() == "/" {
			file, err = wikiio.Load("/Home")
		} else {
			wpath := filepath.Join(page.FilePath(), filepath.Base(page.FilePath()))
			file, err = wikiio.Load(wpath)
		}
		if err != nil {
			LoggerV.Println("main.Page.raw: main file not found")
			return nil
		}
		return file.Raw()
	} else { // page.filePath isn't dir
		return page.Raw()
	}
}

func (page *Page) MarkdownText() string {
	return string(page.raw())
}

// MarkdownHTML converts markdown text (with wikilink) to html
func (page *Page) MarkdownHTML() template.HTML {
	if page.markdownHTML == "" {
		html := blackfriday.MarkdownCommon(page.raw())

		var dir string
		if page.IsDir() {
			dir = page.WikiPath()
		} else {
			dir = filepath.Dir(page.WikiPath())
		}

		page.markdownHTML = template.HTML(processWikiLink(html, dir))
	}

	return page.markdownHTML
}

func (page *Page) SidebarHTML() template.HTML {
	path := filepath.Join(setting.WikiRoot, SIDEBAR_MD)
	_, err := os.Stat(path)
	if err != nil {
		LoggerV.Println(SIDEBAR_MD, "not found")
		return template.HTML("")
	}

	LoggerV.Println(SIDEBAR_MD, "found")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		LoggerE.Println("in SidebarHTML ioutil.ReadFile(", path, ") error", err)
		return template.HTML("")
	}

	html := blackfriday.MarkdownCommon(b)
	return template.HTML(processWikiLink(html, "/"))
}

func (page *Page) Render(res http.ResponseWriter) error {
	return views.ExecuteTemplate(res, PAGE_VIEW, &page)
}
