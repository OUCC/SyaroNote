package main

import (
	"github.com/OUCC/SyaroNote/syaro/markdown"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzers/keyword_analyzer"
	_ "github.com/blevesearch/bleve/analysis/language/en"
	// _ "github.com/blevesearch/bleve/analysis/language/ja"

	"gopkg.in/fsnotify.v1"

	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

const (
	BLEVE_PATH = ".syaronote.bleve"
)

var (
	bleveIndex bleve.Index

	updateIndex = make(chan fsnotify.Event)
)

type wikiPageIndex struct {
	Name string `json:"name"`

	Title   string   `json:"title"`
	Author  string   `json:"author"`
	Aliases []string `json:"aliases"`
	Tags    []string `json:"tags"`

	Contents string `json:"contents"`
}

// implement bleve.Classifier
func (w *wikiPageIndex) Type() string {
	return "wiki"
}

// must be called after setting.wikiRoot is set
func indexLoop() {
	blevePath := filepath.Join(setting.wikiRoot, BLEVE_PATH)
	os.RemoveAll(blevePath)
	mapping, err := buildIndexMapping()
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}
	bleveIndex, err = bleve.New(blevePath, mapping)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	log.Info("Building index...")
	batch := bleveIndex.NewBatch()
	err = filepath.Walk(setting.wikiRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		wpath := toWikiPath(path)
		data, err := loadPageIndex(wpath)
		if err != nil {
			return nil
		}
		log.Debug("Indexing %s", wpath)
		return batch.Index(wpath, data)
	})
	if err != nil || bleveIndex.Batch(batch) != nil {
		log.Fatalf("Failed to create index: %v", err)
	}
	batch.Reset()
	log.Info("Index building end")

	// loop for updating index
	for {
		ev := <-updateIndex
		wpath := toWikiPath(ev.Name)

		switch ev.Op {
		case fsnotify.Create, fsnotify.Write:
			data, err := loadPageIndex(wpath)
			if err != nil {
				continue
			}
			log.Debug("Index %s", wpath)
			if bleveIndex.Index(wpath, data) != nil {
				log.Debug("Indexing falied: %v", err)
			}

		case fsnotify.Rename, fsnotify.Remove:
			log.Debug("Delete %s", wpath)
			if bleveIndex.Delete(wpath) != nil {
				log.Debug("Indexing falied: %v", err)
			}
		}
	}
}

func buildIndexMapping() (*bleve.IndexMapping, error) {
	log.Info("Using indexing mode %s", setting.IndexingMode)

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = setting.IndexingMode

	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword_analyzer.Name
	keywordFieldMapping.IncludeInAll = false
	// keywordFieldMapping.IncludeTermVectors = false

	wikiMapping := bleve.NewDocumentStaticMapping()
	wikiMapping.AddFieldMappingsAt("name", keywordFieldMapping)
	wikiMapping.AddFieldMappingsAt("title", textFieldMapping)
	wikiMapping.AddFieldMappingsAt("author", keywordFieldMapping)
	wikiMapping.AddFieldMappingsAt("aliases", keywordFieldMapping)
	wikiMapping.AddFieldMappingsAt("tags", keywordFieldMapping)
	wikiMapping.AddFieldMappingsAt("contents", textFieldMapping)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultAnalyzer = setting.IndexingMode
	mapping.AddDocumentMapping("wiki", wikiMapping)

	for _, s := range []string{"name", "title", "contents"} {
		log.Debug("Field: %s, Analyzer: %s", s, mapping.FieldAnalyzer(s))
	}

	return mapping, nil
}

func loadPageIndex(wpath string) (*wikiPageIndex, error) {
	if strings.Contains(wpath, string(filepath.Separator)+".") || // exclude hidden files
		strings.HasSuffix(wpath, FOLDER_MD) { // exclude _.md
		return nil, fmt.Errorf("excluded file")
	}

	wf, err := loadFile(wpath)
	if err != nil {
		return nil, err
	}

	var name string
	if wpath == string(filepath.Separator) {
		name = wpath
	} else {
		name = removeExt(wf.Name())
	}

	var b []byte
	switch wf.fileType {
	case WIKIFILE_MARKDOWN:
		var err error
		b, err = wf.read()
		if err != nil {
			return nil, err
		}
	case WIKIFILE_FOLDER:
		wf_, err := loadFile(filepath.Join(wf.WikiPath, FOLDER_MD))
		if err != nil {
			return &wikiPageIndex{
				Name: name,
			}, nil
		}
		if b, err = wf_.read(); err != nil {
			return nil, err
		}
	case WIKIFILE_OTHER:
		return nil, fmt.Errorf("not a markdown")
	}

	meta := markdown.Meta(b)

	return &wikiPageIndex{
		Name:     name,
		Title:    html.EscapeString(meta["title"]),
		Author:   meta["author"],
		Aliases:  splitCommma(meta["alias"] + "," + meta["aliases"]),
		Tags:     splitCommma(meta["tag"] + "," + meta["tags"]),
		Contents: html.EscapeString(string(b)),
	}, nil
}
