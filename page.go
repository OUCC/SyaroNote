package syaro

import (
	"bufio"
	"bytes"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
)

type Page struct {
	FilePath     string
	Title        string
	MarkdownText string
	MarkdownHTML template.HTML
}

func LoadPage(filePath string) (*Page, error) {
	// open md file
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	// read md file
	md, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// convert md to html
	mdhtml := blackfriday.MarkdownCommon(md)

	// find title
	title := searchTitle(mdhtml)

	return &Page{
		FilePath:     filePath,
		Title:        title,
		MarkdownText: string(md),
		MarkdownHTML: template.HTML(mdhtml),
	}, nil
}

func NewPage(filePath string) (*Page, error) {
	// create md file
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	return &Page{
		FilePath:     filePath,
		Title:        "",
		MarkdownText: "",
		MarkdownHTML: template.HTML(""),
	}, nil
}

func (page *Page) Render(rw http.ResponseWriter) error {
	// read template html
	filePath := path.Clean(templateDir + "/page.html")
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	html, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// parce html
	tmpl, err := template.New(page.Title).Parse(string(html))
	if err != nil {
		return err
	}

	// render
	return tmpl.Execute(rw, &page)
}

func (page *Page) Save(mdtext []byte) error {
	// open md file
	f, err := os.Open(page.FilePath)
	defer f.Close()
	if err != nil {
		return err
	}

	// delete all content
	err = f.Truncate(0)
	if err != nil {
		return err
	}

	// write
	_, err = f.Write(mdtext)
	if err != nil {
		return err
	}

	// convert md to html
	page.MarkdownHTML = template.HTML(blackfriday.MarkdownCommon(mdtext))

	return nil
}

func searchTitle(b []byte) string {
	reader := bytes.NewReader(b)
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
