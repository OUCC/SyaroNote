package main

import (
	"github.com/blevesearch/bleve"

	"html/template"
	"net/http"
	"time"
)

const (
	SEARCH_TMPL = "search.html"
)

type searchData struct {
	Query   string
	Total   int
	Results []searchResultData
	Took    time.Duration
	Sidebar template.HTML
}

type searchResultData struct {
	WikiPath  string
	Name      string
	Score     float64
	Title     template.HTML
	Fragments []template.HTML
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	log.Info("Searching... q: %s", q)
	query := bleve.NewQueryStringQuery(q)
	request := bleve.NewSearchRequest(query)
	request.Highlight = bleve.NewHighlight()
	request.Highlight.AddField("title")
	request.Highlight.AddField("contents")
	request.Fields = []string{"name"}
	result, err := bleveIndex.Search(request)
	if err != nil {
		log.Error("Failed to search: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Debug(result.String())

	log.Info("Rendering search page...")
	if err := renderSearch(w, q, result); err != nil {
		log.Error("Rendering error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("OK")
}

func renderSearch(w http.ResponseWriter, q string, result *bleve.SearchResult) error {
	results := make([]searchResultData, len(result.Hits))
	for i, hit := range result.Hits {
		var fragments []template.HTML
		if frags, ok := hit.Fragments["contents"]; ok {
			fragments = make([]template.HTML, len(frags))
			for j, frag := range frags {
				fragments[j] = template.HTML(frag)
			}
		}
		var title template.HTML
		if frags, ok := hit.Fragments["title"]; ok && len(frags) > 0 {
			title = template.HTML(frags[0])
		}
		results[i] = searchResultData{
			WikiPath:  hit.ID,
			Name:      hit.Fields["name"].(string),
			Score:     hit.Score,
			Title:     title,
			Fragments: fragments,
		}
	}

	data := searchData{
		Query:   q,
		Results: results,
		Took:    result.Took,
		Sidebar: loadSidebar(),
	}
	return tmpl.ExecuteTemplate(w, SEARCH_TMPL, data)
}
