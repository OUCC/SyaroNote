package main

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/util"
	"github.com/OUCC/syaro/wikiio"

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
	*wikiio.WikiFile
	mdHTML template.HTML
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
		ret[i], _ = LoadPage("/" + strings.Join(s[:i+1], "/"))
	}
	return ret
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
		LoggerV.Println("main.Page.raw: requested page is dir, use main file of dir")

		if page.DirMainPage() != nil {
			LoggerV.Println("main.Page.raw: main file of dir found")
			return page.DirMainPage().Raw()
		} else {
			LoggerV.Println("main.Page.raw: main file of dir not found")
			return nil
		}
	} else { // page.filePath isn't dir
		return page.Raw()
	}
}

func (page *Page) MdText() string {
	return string(page.raw())
}

func (page *Page) MdHTML() string {
	// if page.mdHTML == "" {
	var dir string
	if page.IsDir() {
		dir = page.WikiPath()
	} else {
		dir = filepath.Dir(page.WikiPath())
	}

	return processWikiLink(page.MdText(), dir)
	// page.mdHTML = template.HTML(mdhtml)
	// }
	// return page.mdHTML
}

func (page *Page) SidebarMdHTML() string {
	path := filepath.Join(setting.WikiRoot, SIDEBAR_MD)
	_, err := os.Stat(path)
	if err != nil {
		LoggerV.Println(SIDEBAR_MD, "not found")
		// return template.HTML("")
		return ""
	}

	LoggerV.Println(SIDEBAR_MD, "found")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		LoggerE.Println("in SidebarHTML ioutil.ReadFile(", path, ") error", err)
		// return template.HTML("")
		return ""
	}

	return processWikiLink(string(b), "/")
	// mdhtml := processWikiLink(html.EscapeString(string(b)), "/")
	// return template.HTML(mdhtml)
}

func (page *Page) Render(res http.ResponseWriter) error {
	return views.ExecuteTemplate(res, PAGE_VIEW, &page)
}
