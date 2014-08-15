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
		loggerV.Println(path, "is out of", setting.wikiRoot)
		return nil, errors.New("requested file is out of setting.wikiRoot")
	}
	if !(path == "/" || path == "" || path == ".") && filepath.Ext(path) != "" && !isMarkdown(path) {
		loggerV.Println(path, "is not a markdown. ignored.")
		return nil, errors.New("requested file is not a markdown")
	}

	_, err := os.Stat(path)
	if err == nil {
		loggerV.Println(path, "found")
		return &Page{filePath: path}, nil

	} else if _, ok := err.(*os.PathError); ok {
		loggerV.Println(path, "not found")
		paths := addExt(path)
		loggerV.Println("search result:", paths)

		if len(paths) == 1 { // only one page found
			loggerV.Println("using", paths[0])
			return &Page{filePath: paths[0]}, nil

		} else if len(paths) > 1 { // more than one page found
			// TODO avoid ambiguous page
			loggerV.Println("more than one page found")
			return nil, errors.New("More than one page found")

		} else { // no page found
			loggerV.Println("no page found")
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
// ex) FilePath: /path/to/wikiroot/hoge/fuga.md
//     WikiPath: /hoge/fuga.md
func (page *Page) WikiPath() string {
	ret, err := filepath.Rel(setting.wikiRoot, page.filePath)
	if err != nil {
		loggerV.Printf("in Page.WikiPath() filepath.Rel(%s, %s) returned error %v",
			setting.wikiRoot, page.filePath, err)
		return ""
	}

	return ret
}

// WikiPathList returns slice of each pages in wikipath
// (slice dosen't include urlPrefix)
func (page *Page) WikiPathList() []*Page {
	s := strings.Split(removeExt(page.WikiPath()), "/")
	if s[0] == "" {
		s = s[1:]
	}

	ret := make([]*Page, len(s))
	for i := 0; i < len(ret); i++ {
		ret[i], _ = LoadPage(filepath.Join(setting.wikiRoot, strings.Join(s[:i+1], "/")))
	}
	return ret
}

// IsDir returns whether path is dir or not.
func (page *Page) IsDir() bool {
	info, err := os.Stat(page.filePath)
	if err != nil {
		loggerE.Panicln(page.filePath, "not found!")
	}
	return info.IsDir()
}

func (page *Page) PageList() []*Page {
	if !page.IsDir() {
		return nil
	}

	loggerV.Println("reading directory", page.filePath, "...")
	infos, err := ioutil.ReadDir(page.filePath)
	if err != nil {
		loggerV.Println("in Page.PageList() ioutil.ReadDir(%s) returned error %v",
			page.filePath, err)
		return nil
	}

	loggerV.Println(len(infos), "file/dirs found")
	ret := make([]*Page, len(infos))
	i := 0
	for _, info := range infos {
		if info.Name()[:1] != "." { // not a hidden file
			ret[i], _ = LoadPage(filepath.Join(page.filePath, info.Name()))
			i++
		}
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
		loggerV.Println("requested page is dir")

		path = filepath.Join(page.filePath, filepath.Base(page.filePath))
		loggerV.Println("searching page", path, "...")
		paths := addExt(path)
		loggerV.Println("search result:", paths)

		if len(paths) == 1 { // only one page found
			loggerV.Println("using", paths[0])
			path = paths[0]

		} else if len(paths) > 1 { // more than one page found
			// TODO avoid ambiguous page
			loggerV.Println("more than one file found")
			loggerV.Println("using", paths[0])
			path = paths[0]

		} else { // no page found
			loggerV.Println("no file found")
			return nil
		}
	} else { // page.filePath isn't dir
		path = page.filePath
	}

	// read md file
	b, err := ioutil.ReadFile(path)
	if err != nil {
		loggerE.Panicln(page.filePath, "not found!")
	}
	return b
}

func (page *Page) MarkdownText() string {
	return string(page.row())
}

// MarkdownHTML converts markdown text (with wikilink) to html
func (page *Page) MarkdownHTML() template.HTML {
	html := blackfriday.MarkdownCommon(page.row())
	return template.HTML(processWikiLink(html, filepath.Dir(page.filePath)))
}

func (page *Page) SidebarHTML() template.HTML {
	path := filepath.Join(setting.wikiRoot, SIDEBAR_MD)
	_, err := os.Stat(path)
	if err != nil {
		loggerV.Println(SIDEBAR_MD, "not found")
		return template.HTML("")
	}

	loggerV.Println(SIDEBAR_MD, "found")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		loggerE.Println("in SidebarHTML ioutil.ReadFile(", path, ") error", err)
		return template.HTML("")
	}

	html := blackfriday.MarkdownCommon(b)
	return template.HTML(processWikiLink(html, filepath.Dir(page.filePath)))
}

func (page *Page) Render(res http.ResponseWriter) error {
	return views.ExecuteTemplate(res, "page.html", &page)
}

func (page *Page) Save(b []byte) error {
	return ioutil.WriteFile(page.FilePath(), b, 0644)
}
