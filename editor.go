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
	LoggerV.Printf("main.NewEditor(%s)\n", wpath)

	page, err := LoadPage(wpath)
	if err != nil {
		LoggerV.Println("Error: main.NewEditor:", err)
		return nil, err
	}

	// check if is dir
	if page.IsDir() {
		LoggerV.Println("Error: main.NewEditor: this is dir")
		return nil, ErrIsNotMarkdown
	}

	LoggerV.Println("main.NewEditor: ok")
	return &Editor{page}, nil
}

func (editor *Editor) Render(res http.ResponseWriter) error {
	return views.ExecuteTemplate(res, EDITOR_VIEW, editor)
}
