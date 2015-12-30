package main

import (
	"github.com/blevesearch/bleve"

	"bytes"
	"strings"
)

func linkWorker(out *bytes.Buffer, b []byte) {
	link := string(b)
	found := resolveLink(link)
	var href string

	if len(found) != 0 { // page found
		// TODO disambiguation page
		log.Debug("%s -> %s", link, found[0])

		href = queryEscapeWikiPath(found[0])
		out.WriteString(`<a href="` + href + `">` + link + `</a>`)
	} else { // page not found
		log.Debug("%s -> NOTFOUND", link)
		href = queryEscapeWikiPath(link)
		out.WriteString(`<a class="notfound" href="` + href + `">` + link + `</a>`)
	}
}

func resolveLink(link string) []string {
	if strings.Contains(link, "/") || isMarkdown(link) {
		// absolute and relative path not supported
		return nil
	}
	log.Debug("Searching name... term: %s", link)
	query := bleve.NewTermQuery(link)
	query.SetField("name")
	request := bleve.NewSearchRequest(query)
	request.Fields = []string{"name"}
	result, err := bleveIndex.Search(request)
	if err != nil {
		log.Debug("Error: %v", err)
	}

	log.Debug(result.String())
	if result.Total > 0 {
		links := make([]string, len(result.Hits))
		for i, hit := range result.Hits {
			links[i] = hit.ID
		}
		return links
	}

	log.Debug("Searching aliases... term: %s", link)
	query.SetField("aliases")
	request = bleve.NewSearchRequest(query)
	request.Fields = []string{"aliases"}
	result, err = bleveIndex.Search(request)
	if err != nil {
		log.Debug("Error: %v", err)
	}

	log.Debug(result.String())
	if result.Total > 0 {
		log.Debug(result.String())
		links := make([]string, len(result.Hits))
		for i, hit := range result.Hits {
			links[i] = hit.ID
		}
		return links
	}
	return nil
}
