package main

import (
	"strings"
)

func linkWorker(link, dir string) string {
	log.Debug("link: %s, dir: %s", link, dir)

	found := resolveLink(link, dir)

	if len(found) != 0 { // page found
		// TODO disambiguation page
		log.Debug("%d pages found", len(found))
		log.Debug("select %s", found[0])

		href := queryEscapeWikiPath(found[0])
		return `<a href="` + href + `">` + link + `</a>`
	} else { // page not found
		log.Debug("no page found")
		href := queryEscapeWikiPath(link)
		return `<a class="notfound" href="` + href + `">` + link + `</a>`
	}
}

func resolveLink(link, dir string) []string {
	abs := filepath.IsAbs(link)
	rel := strings.Contains(link, "/") || isMarkdown(link)
	if abs || rel {
		var wpath string
		if abs {
			// search name as absolute path
			// example: /piyo /poyo/pyon.ext
			wpath = link
		} else {
			// search name as relative path
			// example: ./hoge ../fuga.ext puyo.ext
			wpath = filepath.Join(dir, link)
		}
		if _, err := loadFile(wpath); err == nil {
			return []string{wpath}
		} else {
			return nil
		}
	} else {
		// search name as base name
		// example: abc
		return nameIdx[link]
	}
}
