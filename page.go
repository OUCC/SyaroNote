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
func LoadPage(path string) (*Page, error) {
	// security check
	if !isIn(path, wikiRoot) { // path is out of wikiRoot
		logger.Println(path, "is out of", wikiRoot)
		return nil, errors.New("requested file is out of wikiRoot")
	}
	if filepath.Ext(path) != "" && !isMarkdown(path) {
		logger.Println(path, "is not a markdown. ignored.")
		return nil, errors.New("requested file is not a markdown")
	}

	_, err := os.Stat(path)
	if err == nil {
		logger.Println(path, "found")
		return &Page{filePath: path}, nil

	} else if _, ok := err.(*os.PathError); ok {
		logger.Println(path, "not found")
		paths := addExt(path)
		logger.Println("search result:", paths)

		if len(paths) == 1 { // only one page found
			return &Page{filePath: paths[0]}, nil

		} else if len(paths) > 1 { // more than one page found
			// TODO avoid ambiguous page
			return nil, errors.New("More than one page found")

		} else { // no page found
			return nil, os.ErrNotExist
		}

	} else {
		return nil, err
	}
}

// FilePath returns file path.
func (page *Page) FilePath() string { return page.filePath }

// IsDir returns whether path is dir or not.
func (page *Page) IsDir() bool {
	info, err := os.Stat(page.filePath)
	if err != nil {
		logger.Panicln(page.filePath, "not found!")
		panic(err)
	}
	return info.IsDir()
}

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
		logger.Panicln(page.filePath, "not found!")
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
