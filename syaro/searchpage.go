package main

import (
	"github.com/blevesearch/bleve"

	"html/template"
	"net/http"
	"strings"
	"time"
)

const (
	SEARCH_TMPL = "search.html"
)

type wikiSearchPage struct {
	Query   string
	Total   int
	Results []searchResult
	// Took    float64
	Took    time.Duration
	Sidebar template.HTML
}

type searchResult struct {
	WikiPath string
	Fragment template.HTML
}

func renderSearchPage(w http.ResponseWriter, q string, result *bleve.SearchResult) error {
	results := make([]searchResult, 0, len(result.Hits))
	for _, hit := range result.Hits {
		var fragment string
		if s, ok := hit.Fragments["contents"]; ok {
			fragment = strings.Join(s, "\n")
		} else if s, ok := hit.Fragments["title"]; ok {
			fragment = strings.Join(s, "\n")
		}
		results = append(results, searchResult{
			WikiPath: hit.ID,
			Fragment: template.HTML(fragment),
		})
	}

	data := wikiSearchPage{
		Query:   q,
		Total:   len(results),
		Results: results,
		// Took:    result.Took.Seconds(),
		Took:    result.Took,
		Sidebar: loadSidebar(),
	}
	return tmpl.ExecuteTemplate(w, SEARCH_TMPL, data)
}
