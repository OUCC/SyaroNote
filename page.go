package main

import (
	"bufio"
	"errors"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Page stores both row markdown and converted html.
type Page struct {
	filePath string
}

// LoadPage returns new Page.
func LoadPage(mdpath string) (*Page, error) {
	info, err := os.Stat(mdpath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		// FIXME return special page
		return nil, errors.New("this is directory, not a file.")
	}

	return &Page{filePath: mdpath}, nil
}

// NewPage create new markdown file in local repo.
func NewPage(mdpath string) (*Page, error) {
	// create md file
	f, err := os.Create(mdpath)
	if err != nil {
		return nil, err
	}
	f.Close()

	return &Page{filePath: mdpath}, nil
}

// FilePath returns file path.
func (page *Page) FilePath() string { return page.filePath }

// FIXME serious performance
// Title returns title of page.
func (page *Page) Title() string {
	reader := strings.NewReader(string(page.MarkdownHTML()))
	scanner := bufio.NewScanner(reader)

	re := regexp.MustCompile("^<h1>([^<]*)</h1>$")
	for scanner.Scan() {
		submatch := re.FindStringSubmatch(scanner.Text())
		if len(submatch) != 0 {
			return submatch[1]
		}
	}
	return ""
}

// FIXME wrong performance (returning value)
// row returns row file data.
func (page *Page) row() []byte {
	// read md file
	b, err := ioutil.ReadFile(page.FilePath())
	if err != nil {
		panic(err)
	}
	return b
}

// FIXME wrong performance (returning value)
func (page *Page) MarkdownText() string {
	return string(page.row())
}

// FIXME wrong performance (returning value)
// MarkdownHTML converts markdown text (with wikilink) to html
func (page *Page) MarkdownHTML() template.HTML {
	html := blackfriday.MarkdownCommon(page.row())
	return template.HTML(processWikiLink(html, filepath.Dir(page.FilePath())))
}

func (page *Page) Render(rw http.ResponseWriter) error {
	// read template html
	html, err := ioutil.ReadFile(filepath.Join(templateDir, "page.html"))
	if err != nil {
		return err
	}

	// parce html
	tmpl, err := template.New(page.Title()).Parse(string(html))
	if err != nil {
		return err
	}

	// render
	return tmpl.Execute(rw, &page)
}

func (page *Page) Save(b []byte) error {
	return ioutil.WriteFile(page.FilePath(), b, 0644)
}
