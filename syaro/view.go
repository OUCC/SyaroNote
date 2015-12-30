package main

import (
	"github.com/blevesearch/bleve"

	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	PAGE_TMPL = "page.html"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	// url unescape (+ -> <Space>)
	wpath := strings.Replace(r.URL.Path, "+", " ", -1)
	// wpath := strings.TrimPrefix(r.URL.Path, setting.urlPrefix)

	wf, err := loadFile(wpath)
	if os.IsNotExist(err) {
		log.Error(err.Error())
		renderError(w, wpath, http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(err.Error())
		renderError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch wf.fileType {
	case WIKIFILE_MARKDOWN, WIKIFILE_FOLDER:
		renderPage(w, wf)
	case WIKIFILE_OTHER:
		sendFile(w, wf)
	}
}

func renderPage(w http.ResponseWriter, wf WikiFile) {
	log.Info("Rendering page (%s)...", wf.WikiPath)
	wp, err := loadPage(wf)
	if err != nil {
		log.Error(err.Error())
		renderError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, PAGE_TMPL, &wp); err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("OK")
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	log.Info("Searching... q: %s", q)
	query := bleve.NewQueryStringQuery(q)
	request := bleve.NewSearchRequest(query)
	request.Highlight = bleve.NewHighlight()
	request.Highlight.AddField("title")
	request.Highlight.AddField("contents")
	result, err := bleveIndex.Search(request)
	if err != nil {
		log.Error("Failed to search: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Debug(result.String())

	log.Info("Rendering search page...")
	if err := renderSearchPage(w, q, result); err != nil {
		log.Error("Rendering error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("OK")
}

func renderError(w http.ResponseWriter, data string, status int) {
	log.Error("Rendering error view... (status: %d, data: %s)", status, data)

	w.WriteHeader(status)
	err := tmpl.ExecuteTemplate(w, strconv.Itoa(status)+".html", data)
	if err != nil {
		log.Error(err.Error())
		w.Write(nil)
	}
}

func sendFile(w http.ResponseWriter, wf WikiFile) {
	log.Info("Sending file (%s)...", wf.WikiPath)
	b, err := wf.read()
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
	log.Info("OK")
}
