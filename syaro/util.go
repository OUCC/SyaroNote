package main

import (
	"container/list"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	// markdown file's extensiton list
	mdExtList = []string{
		".md",
		".mkd",
		".mkdn",
		".mdown",
		".markdown",
	}
)

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

func addExt(pathWithoutExt string) []string {
	files := list.New()

	for _, ext := range mdExtList {
		path := pathWithoutExt + ext
		_, err := os.Stat(path)
		if err == nil { // file exists
			files.PushBack(path)
		}
	}

	return toStringArray(files)
}

func toStringArray(src *list.List) []string {
	ret := make([]string, src.Len())
	for i, v := 0, src.Front(); i < len(ret); i, v = i+1, v.Next() {
		ret[i] = v.Value.(string)
	}
	return ret
}

// isIn returns true when dirA is in dirB
func isIn(dirA, dirB string) bool {
	dirA, _ = filepath.Abs(dirA)
	dirB, _ = filepath.Abs(dirB)

	return strings.HasPrefix(dirA, dirB)
}

func queryEscapeWikiPath(wpath string) string {
	// replace all separator to /
	s := strings.Replace(wpath, string(filepath.Separator), "/", -1)
	// url escape
	s = url.QueryEscape(wpath)
	// replace all %2F to /
	s = strings.Replace(s, "%2F", "/", -1)
	// replace all + to %20
	s = strings.Replace(s, "+", "%20", -1)
	return s
}

func toWikiPath(path string) string {
	wpath, err := filepath.Rel(setting.wikiRoot, path)
	if err != nil || strings.HasPrefix(wpath, "..") {
		return "/"
	}
	return filepath.Join(string(filepath.Separator), wpath)
}
