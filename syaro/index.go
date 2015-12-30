package main

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzers/custom_analyzer"
	"github.com/blevesearch/bleve/analysis/analyzers/keyword_analyzer"
	"github.com/blevesearch/bleve/analysis/char_filters/html_char_filter"
	"github.com/blevesearch/bleve/analysis/language/en"
	"github.com/blevesearch/bleve/analysis/language/ja"
	"github.com/blevesearch/bleve/analysis/token_filters/lower_case_filter"
	"github.com/blevesearch/bleve/analysis/token_filters/porter"
	"github.com/blevesearch/bleve/analysis/token_filters/unicode_normalize"
	"github.com/blevesearch/bleve/analysis/tokenizers/unicode"

	"gopkg.in/fsnotify.v1"

	"os"
	"path/filepath"
	// "regexp"
	"strings"
)

const (
	BLEVE_PATH = ".syaronote.bleve"
)

var (
	bleveIndex bleve.Index

	updateIndex = make(chan fsnotify.Event)
)

// must be called after setting.wikiRoot is set
func idxBuilder() {
	blevePath := filepath.Join(setting.wikiRoot, BLEVE_PATH)
	os.RemoveAll(blevePath)
	mapping, err := buildIndexMapping()
	if err != nil {
		log.Fatal("Failed to create index: %v", err)
	}
	bleveIndex, err = bleve.New(blevePath, mapping)
	if err != nil {
		log.Fatal("Failed to create index: %v", err)
	}

	log.Info("Building index...")
	batch := bleveIndex.NewBatch()
	err = filepath.Walk(setting.wikiRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// exclude hidden files
		if strings.Contains(path, "/.") {
			return nil
		}
		wpath := toWikiPath(path)
		wf, err := loadFile(wpath)
		if err != nil || wf.fileType == WIKIFILE_OTHER {
			return nil
		}
		wp, err := loadPage(wf)
		if err != nil {
			return nil
		}
		log.Debug("Indexing %s", wpath)
		return batch.Index(wpath, wp)
	})
	if err != nil || bleveIndex.Batch(batch) != nil {
		log.Fatal("Failed to create index: %v", err)
	}
	log.Info("Index building end")

	// loop for updating index
	for {
		ev := <-updateIndex
		wpath := toWikiPath(ev.Name)
		switch ev.Op {
		case fsnotify.Create, fsnotify.Write:
			log.Debug("Index %s", wpath)
			wf, err := loadFile(wpath)
			if err != nil || wf.fileType == WIKIFILE_OTHER {
				continue
			}
			wp, err := loadPage(wf)
			if err != nil {
				continue
			}
			if bleveIndex.Index(wpath, wp) != nil {
				log.Debug("Indexing falied: %v", err)
			}

		case fsnotify.Rename, fsnotify.Remove:
			log.Debug("Delete %s", wpath)
			if bleveIndex.Delete(wpath) != nil {
				log.Debug("Indexing falied: %v", err)
			}
		}
	}

	// 		case q := <-searchName:
	// 			log.Debug("Searching name... q: %s", q)
	// 			query := bleve.NewPhraseQuery([]string{q}, "name")
	// 			request := bleve.NewSearchRequest(query)
	// 			result, err := bleveIndex.Search(request)
	// 			if err != nil {
	// 				log.Debug("Error: %v", err)
	// 			} else if result.Total > 0 {
	// 				log.Debug(result.String())
	// 				continue
	// 			}

	// 			query = bleve.NewPhraseQuery([]string{q}, "aliases")
	// 			request = bleve.NewSearchRequest(query)
	// 			result, err = bleveIndex.Search(request)
	// 			if err != nil {
	// 				log.Debug("Error: %v", err)
	// 			} else {
	// 				log.Debug(result.String())
	// 			}

	// 		case q := <-searchText:
	// 			log.Debug("Searching text... q: %s", q)
	// 			query := bleve.NewQueryStringQuery(q)
	// 			request := bleve.NewSearchRequest(query)
	// 			request.Highlight = bleve.NewHighlightWithStyle(ansi.Name)
	// 			request.Highlight.AddField("contents")
	// 			result, err := bleveIndex.Search(request)
	// 			if err != nil {
	// 				log.Debug("Error: %v", err)
	// 			} else {
	// 				log.Debug(result.String())
	// 			}
	// 		}
	// }
}

func buildIndexMapping() (*bleve.IndexMapping, error) {
	// reAnalizer := regexp.MustCompile(`^standard|en|ja$`)
	analizer := "standard"
	// if reAnalizer.MatchString(setting.IndexingMode) {
	// analizer = setting.IndexingMode
	// }
	log.Info("Using indexing mode %s", analizer)

	titleFieldMapping := bleve.NewTextFieldMapping()
	titleFieldMapping.Analyzer = analizer

	htmlFieldMapping := bleve.NewTextFieldMapping()
	htmlFieldMapping.Analyzer = analizer + "html"

	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword_analyzer.Name
	keywordFieldMapping.IncludeInAll = false
	keywordFieldMapping.IncludeTermVectors = false

	wikiMapping := bleve.NewDocumentStaticMapping()
	wikiMapping.AddFieldMappingsAt("name", keywordFieldMapping)
	wikiMapping.AddFieldMappingsAt("title", titleFieldMapping)
	wikiMapping.AddFieldMappingsAt("author", keywordFieldMapping)
	wikiMapping.AddFieldMappingsAt("aliases", keywordFieldMapping)
	wikiMapping.AddFieldMappingsAt("tags", keywordFieldMapping)
	wikiMapping.AddFieldMappingsAt("contents", htmlFieldMapping)

	mapping := bleve.NewIndexMapping()
	err := mapping.AddCustomAnalyzer("standardhtml", map[string]interface{}{
		"type": custom_analyzer.Name,
		"char_filters": []string{
			html_char_filter.Name,
		},
		"tokenizer": unicode.Name,
		"token_filters": []string{
			lower_case_filter.Name,
			en.StopName,
		},
	})
	if err != nil {
		return nil, err
	}
	err = mapping.AddCustomTokenFilter(unicode_normalize.NFKD, map[string]interface{}{
		"type": unicode_normalize.Name,
		"form": unicode_normalize.NFKD,
	})
	if err != nil {
		return nil, err
	}
	err = mapping.AddCustomAnalyzer("jahtml", map[string]interface{}{
		"type": custom_analyzer.Name,
		"char_filters": []string{
			html_char_filter.Name,
		},
		"tokenizer": ja.TokenizerName,
		"token_filters": []string{
			unicode_normalize.NFKD,
		},
	})
	if err != nil {
		return nil, err
	}
	err = mapping.AddCustomAnalyzer("enhtml", map[string]interface{}{
		"type": custom_analyzer.Name,
		"char_filters": []string{
			html_char_filter.Name,
		},
		"tokenizer": unicode.Name,
		"token_filters": []string{
			lower_case_filter.Name,
			en.StopName,
			en.PossessiveName,
			porter.Name,
		},
	})
	if err != nil {
		return nil, err
	}
	mapping.DefaultAnalyzer = analizer
	// mapping.DefaultField = "body"
	mapping.AddDocumentMapping("wiki", wikiMapping)

	for _, s := range []string{"name", "title", "contents"} {
		log.Debug("Field: %s, Analyzer: %s", s, mapping.FieldAnalyzer(s))
	}

	return mapping, nil
}
