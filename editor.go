package main

import (
	"net/http"
)

const (
	EDITOR_VIEW = "editor.html"
)

type Editor struct {
	*Page
}

func NewEditor(wpath string) (*Editor, error) {
	log.Debug("wpath: %s", wpath)

	page, err := LoadPage(wpath)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	// check if is dir
	if page.IsDir() {
		log.Error("this is dir")
		return nil, ErrIsNotMarkdown
	}

	log.Debug("ok")
	return &Editor{page}, nil
}

func (editor *Editor) Render(res http.ResponseWriter) error {
	return views.ExecuteTemplate(res, EDITOR_VIEW, editor)
}
