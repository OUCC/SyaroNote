package main

import (
	"github.com/OUCC/syaro/markdown"

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

	Contents   template.HTML
	Sidebar    template.HTML
	TOC        template.HTML
	Title      string
	Tags       []string
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
	var dir string
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
		dir = wf.WikiPath

		// file list
		fis := wf.files()
		wp.Folders = make([]WikiFile, 0, len(fis))
		wp.MdFiles = make([]WikiFile, 0, len(fis))
		wp.OtherFiles = make([]WikiFile, 0, len(fis))
		for _, fi := range fis {
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
		dir = filepath.Dir(wf.WikiPath)
	}

	markdown.LinkWorker = func(b []byte) []byte {
		s := string(b)
		if len(s) < 5 {
			return nil
		}
		link := s[2 : len(s)-2]
		return []byte(linkWorker(link, dir))
	}
	wp.Contents = template.HTML(markdown.Convert(b, dir))
	wp.TOC = template.HTML(markdown.TOC(b))

	// meta datas
	meta := markdown.Meta(b)
	if meta.Title != "" {
		wp.Title = meta.Title
	} else {
		wp.Title = removeExt(wp.Name())
	}
	wp.Tags = meta.Tags

	// breadcrumb list
	bc := make([]WikiFile, 0)
	for wf2, ok := wf.parent(); ok; wf2, ok = wf2.parent() {
		bc = append(bc, wf2)
	}
	// remove wiki root
	if len(bc) > 1 {
		l := len(bc) - 1
		wp.BreadCrumb = make([]WikiFile, l)
		// reverse list
		for i := 0; i < l; i++ {
			wp.BreadCrumb[i] = bc[l-i]
		}
	} else {
		bc = nil
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
	}
	html = template.HTML(markdown.Convert(b, ""))
	return
}
