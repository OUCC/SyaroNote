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

const (
	PAGE_TMPL  = "page.html"
	FLIST_TMPL = "filelist.html"
	SIDEBAR_MD = "_Sidebar.md"
)

// Page stores both row markdown and converted html.
type Page struct {
	filePath string
}

// LoadPage returns new Page.
func LoadPage(path string) (*Page, error) {
	// security check
	if !isIn(path, setting.wikiRoot) { // path is out of setting.wikiRoot
		logger.Println(path, "is out of", setting.wikiRoot)
		return nil, errors.New("requested file is out of setting.wikiRoot")
	}
	if !(path == "/" || path == "" || path == ".") && filepath.Ext(path) != "" && !isMarkdown(path) {
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
			logger.Println("using", paths[0])
			return &Page{filePath: paths[0]}, nil

		} else if len(paths) > 1 { // more than one page found
			// TODO avoid ambiguous page
			logger.Println("more than one page found")
			return nil, errors.New("More than one page found")

		} else { // no page found
			logger.Println("no page found")
			return nil, os.ErrNotExist
		}

	} else {
		return nil, err
	}
}

// Name returns page name
func (page *Page) Name() string {
	return removeExt(filepath.Base(page.filePath))
}

// FilePath returns file path.
func (page *Page) FilePath() string { return page.filePath }

// WikiPath returns file path relative to setting.wikiRoot
func (page *Page) WikiPath() string {
	ret, err := filepath.Rel(setting.wikiRoot, page.filePath)
	if err != nil {
		logger.Println("in Page.WikiPath() filepath.Rel(", setting.wikiRoot, ",", page.filePath,
			") returned error", err)
		return ""
	}

	return "/" + ret
}

// IsDir returns whether path is dir or not.
func (page *Page) IsDir() bool {
	info, err := os.Stat(page.filePath)
	if err != nil {
		logger.Panicln(page.filePath, "not found!")
	}
	return info.IsDir()
}

func (page *Page) PageList() []*Page {
	if !page.IsDir() {
		return nil
	}

	logger.Println("reading directory", page.filePath, "...")
	infos, err := ioutil.ReadDir(page.filePath)
	if err != nil {
		logger.Println("in Page.PageList() ioutil.ReadDir(", page.filePath,
			") returned error", err)
		return nil
	}

	logger.Println(len(infos), "file/dirs found")
	ret := make([]*Page, len(infos))
	for i, info := range infos {
		ret[i], _ = LoadPage(filepath.Join(page.filePath, info.Name()))
	}
	return ret
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

// row returns row file data.
func (page *Page) row() []byte {
	var path string
	if page.IsDir() {
		logger.Println("requested page is dir")

		path = filepath.Join(page.filePath, filepath.Base(page.filePath))
		logger.Println("searching page", path, "...")
		paths := addExt(path)
		logger.Println("search result:", paths)

		if len(paths) == 1 { // only one page found
			logger.Println("using", paths[0])
			path = paths[0]

		} else if len(paths) > 1 { // more than one page found
			// TODO avoid ambiguous page
			logger.Println("more than one file found")
			logger.Println("using", paths[0])
			path = paths[0]

		} else { // no page found
			logger.Println("no file found")
			return nil
		}
	} else { // page.filePath isn't dir
		path = page.filePath
	}

	// read md file
	b, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Panicln(page.filePath, "not found!")
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
	return template.HTML(processWikiLink(html, filepath.Dir(page.filePath)))
}

func (page *Page) SidebarHTML() template.HTML {
	path := filepath.Join(setting.wikiRoot, SIDEBAR_MD)
	_, err := os.Stat(path)
	if err != nil {
		logger.Println(SIDEBAR_MD, "not found")
		return template.HTML("")
	}

	logger.Println(SIDEBAR_MD, "found")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Println("in SidebarHTML ioutil.ReadFile(", path, ") error", err)
		return template.HTML("")
	}

	html := blackfriday.MarkdownCommon(b)
	return template.HTML(processWikiLink(html, filepath.Dir(page.filePath)))
}

func (page *Page) Render(rw http.ResponseWriter) error {
	// read template html
	html, err := ioutil.ReadFile(filepath.Join(setting.tmplDir, PAGE_TMPL))
	if err != nil {
		return err
	}

	// funcs for calculation on template
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}

	// parce html
	tmpl, err := template.New(page.Title()).Funcs(funcMap).Parse(string(html))
	if err != nil {
		return err
	}

	// render
	return tmpl.Execute(rw, &page)
}

func (page *Page) Save(b []byte) error {
	return ioutil.WriteFile(page.FilePath(), b, 0644)
}
