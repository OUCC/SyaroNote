package main

import (
	. "github.com/OUCC/syaro/logger"

	"net/http"
)

const (
	EDITOR_VIEW = "editor.html"
)

type Editor struct {
	*Page
}

func NewEditor(wpath string) (*Editor, error) {
	Log.Debug("wpath: %s", wpath)

	page, err := LoadPage(wpath)
	if err != nil {
		Log.Error(err.Error())
		return nil, err
	}

	// check if is dir
	if page.IsDir() {
		Log.Error("this is dir")
		return nil, ErrIsNotMarkdown
	}

	Log.Debug("ok")
	return &Editor{page}, nil
}

func (editor *Editor) Render(res http.ResponseWriter) error {
	return views.ExecuteTemplate(res, EDITOR_VIEW, editor)
}
