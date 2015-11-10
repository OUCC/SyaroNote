package main

import (
	"strings"
)

func linkWorker(b []byte) []byte {
	link := string(b)
	if len(link) < 5 {
		return nil
	}
	link = link[2 : len(link)-2]
	log.Debug("link: %s", link)

	found := resolveLink(link)

	if len(found) != 0 { // page found
		// TODO disambiguation page
		log.Debug("%d pages found", len(found))
		log.Debug("select %s", found[0])

		href := queryEscapeWikiPath(found[0])
		return []byte(`<a href="` + href + `">` + link + `</a>`)
	} else { // page not found
		log.Debug("no page found")
		href := queryEscapeWikiPath(link)
		return []byte(`<a class="notfound" href="` + href + `">` + link + `</a>`)
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
