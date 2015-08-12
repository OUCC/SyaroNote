package main

import (
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

var (
	reSetext = regexp.MustCompile("^={2,}")
	reAtx    = regexp.MustCompile("^#\\s+([^#]+)")
)

type Page struct {
	*WikiFile
}

// LoadPage returns new Page.
func LoadPage(wpath string) (*Page, error) {
	log.Debug("wpath: %s", wpath)

	wfile, err := loadFile(wpath)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}

	// check if file isn't markdown
	if !(wfile.IsDir() || wfile.IsMarkdown()) {
		log.Debug("file isn't markdown")
		return nil, ErrIsNotMarkdown
	}

	log.Debug("ok")
	return &Page{wfile}, nil
}

// Title returns title of page.
func (page *Page) Title() string {
	reader := strings.NewReader(page.MdText())
	scanner := bufio.NewScanner(reader)

	var previous string
	for scanner.Scan() {
		s := scanner.Text()
		if reSetext.MatchString(s) {
			return previous
		}
		if reAtx.MatchString(s) {
			return reAtx.FindStringSubmatch(s)[1]
		}

		previous = s
	}

	// h1 not found
	return page.NameWithoutExt()
}

// row returns row file data.
func (page *Page) raw() []byte {
	if page.IsDir() {
		log.Debug("requested page is dir, use main file of dir")

		if page.DirMainPage() != nil {
			log.Debug("main file of dir found")
			return page.DirMainPage().Raw()
		} else {
			log.Debug("main file of dir not found")
			return nil
		}
	} else { // page.filePath isn't dir
		return page.Raw()
	}
}

func (page *Page) MdText() string {
	return string(page.raw())
}

func (page *Page) MdHTML() template.HTML {
	text := page.MdText()
	if text == "" {
		return template.HTML("")
	}

	var dir string
	if page.IsDir() {
		dir = page.WikiPath()
	} else {
		dir = filepath.Dir(page.WikiPath())
	}

	return template.HTML(parseMarkdown([]byte(text), dir))
}

func (page *Page) SidebarMdHTML() template.HTML {
	path := filepath.Join(setting.wikiRoot, SIDEBAR_MD)
	_, err := os.Stat(path)
	if err != nil {
		log.Debug("%s not found", SIDEBAR_MD)
		// return template.HTML("")
		return ""
	}

	log.Debug("%s found", SIDEBAR_MD)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("in SidebarHTML ioutil.ReadFile(%s) %s", path, err)
		// return template.HTML("")
		return ""
	}

	return template.HTML(parseMarkdown(b, "/"))
}

func (page *Page) Render(res http.ResponseWriter) error {
	return views.ExecuteTemplate(res, PAGE_VIEW, &page)
}
