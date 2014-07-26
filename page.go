package main

import (
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
)

type Page struct {
	FilePath     string
	MarkDownText []byte
	HTMLBody     []byte
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
	html := blackfriday.MarkdownCommon(md)

	return &Page{
		FilePath:     filePath,
		MarkDownText: md,
		HTMLBody:     html,
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
		MarkDownText: nil,
		HTMLBody:     nil,
	}, nil
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
	page.HTMLBody = blackfriday.MarkdownCommon(mdtext)

	return nil
}
