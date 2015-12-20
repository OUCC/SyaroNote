package main

import (
	"bytes"
	"strings"
)

func linkWorker(out *bytes.Buffer, b []byte) {
	link := string(b)
	log.Debug("link: %s", link)
	found := resolveLink(link)
	var href string

	if len(found) != 0 { // page found
		// TODO disambiguation page
		log.Debug("%d pages found", len(found))
		log.Debug("select %s", found[0])

		href = queryEscapeWikiPath(found[0])
		out.WriteString(`<a href="` + href + `">` + link + `</a>`)
	} else { // page not found
		log.Debug("no page found")
		href = queryEscapeWikiPath(link)
		out.WriteString(`<a class="notfound" href="` + href + `">` + link + `</a>`)
	}
}

func resolveLink(link string) []string {
	if strings.Contains(link, "/") || isMarkdown(link) {
		// absolute and relative path not supported
		return nil
	} else {
		// search name as base name
		// example: abc
		return nameIdx[link]
	}
}
