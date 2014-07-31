package main

import (
	"path/filepath"
	"strings"
)

var (
	mdExtList []string
)

func init() {
	// markdown file's extensiton list
	mdExtList = make([]string, 5)
	mdExtList[0] = ".md"
	mdExtList[1] = ".mkd"
	mdExtList[2] = ".mkdn"
	mdExtList[3] = ".mdown"
	mdExtList[4] = ".markdown"
}

func isMarkdown(filename string) bool {
	ext := filepath.Ext(filename)
	for _, mdext := range mdExtList {
		if ext == mdext {
			return true
		}
	}
	return false
}

// TODO test
func removeExt(filename string) string {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)

	if strings.Contains(base, ".") {
		base = base[:strings.LastIndex(base, ".")]
	}

	return filepath.Join(dir, base)
}
