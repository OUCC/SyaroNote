package main

import (
	"github.com/OUCC/syaro/markdown"
	"strings"

	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	FOLDER_MD  = "_.md"
	SIDEBAR_MD = "_Sidebar.md"
)

type WikiPage struct {
	WikiFile

	Name string

	Meta    map[string]string
	Title   string
	Author  string
	Date    string
	Tags    []string
	MathJax string // TODO

	Contents template.HTML
	Sidebar  template.HTML
	TOC      template.HTML

	IsRoot     bool
	Parent     WikiFile
	BreadCrumb []WikiFile
	Folders    []WikiFile
	MdFiles    []WikiFile
	OtherFiles []WikiFile
}

func loadPage(wf WikiFile) (WikiPage, error) {
	wp := WikiPage{WikiFile: wf}
	wp.Sidebar = loadSidebar()

	var b []byte
	switch wf.fileType {
	case WIKIFILE_FOLDER:
		// load _.md
		wf_, err := loadFile(filepath.Join(wf.WikiPath, FOLDER_MD))
		if err == nil { // _.md found
			b, err = wf_.read()
			if err != nil {
				return wp, err
			}
		}

		// file list
		fis := wf.files()
		wp.Folders = make([]WikiFile, 0, len(fis))
		wp.MdFiles = make([]WikiFile, 0, len(fis))
		wp.OtherFiles = make([]WikiFile, 0, len(fis))
		for _, fi := range fis {
			if strings.HasPrefix(fi.Name(), ".") { // is hidden file/dir
				continue
			}
			switch fi.fileType {
			case WIKIFILE_FOLDER:
				wp.Folders = append(wp.Folders, fi)
			case WIKIFILE_MARKDOWN:
				wp.MdFiles = append(wp.MdFiles, fi)
			case WIKIFILE_OTHER:
				wp.OtherFiles = append(wp.OtherFiles, fi)
			}
		}
	case WIKIFILE_MARKDOWN:
		var err error
		b, err = wf.read()
		if err != nil {
			return wp, err
		}
	}

	wp.Contents = template.HTML(markdown.Convert(b))
	wp.TOC = template.HTML(markdown.TOC(b))

	// meta datas
	wp.Meta = markdown.Meta(b)
	wp.Title = wp.Meta["title"]
	wp.Author = wp.Meta["author"]
	wp.Date = wp.Meta["date"]
	wp.Tags = strings.Split(wp.Meta["tags"], ",")

	var ok bool
	wp.Parent, ok = wf.parent()
	wp.IsRoot = !ok

	if wp.IsRoot {
		wp.Name = "/"
	} else {
		wp.Name = removeExt(wp.WikiFile.Name())
	}

	// breadcrumb list
	var bc []WikiFile
	for wf2, ok := wf.parent(); ok; wf2, ok = wf2.parent() {
		bc = append(bc, wf2)
	}
	// remove wiki root
	if len(bc) > 1 {
		l := len(bc) - 1 // parent count
		wp.BreadCrumb = make([]WikiFile, l+1)
		// reverse list
		for i := 0; i < l; i++ {
			wp.BreadCrumb[i] = bc[l-i-1]
		}
		wp.BreadCrumb[l] = wp.WikiFile // self
	} else if len(bc) == 1 { // file under wiki root
		wp.BreadCrumb = []WikiFile{wp.WikiFile}
	} else {
		wp.BreadCrumb = nil
	}

	return wp, nil
}

func loadSidebar() (html template.HTML) {
	path := filepath.Join(setting.wikiRoot, SIDEBAR_MD)

	if _, err := os.Stat(path); err != nil {
		return // not found
	}

	log.Debug("%s found", SIDEBAR_MD)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err.Error())
		return
	}

	html = template.HTML(markdown.Convert(b))
	return
}
