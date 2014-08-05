package main

import (
	"container/list"
	"os"
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
	dirA, err := filepath.Abs(dirA)
	if err != nil {
		loggerE.Fatalf("filepath.Abs(%v)", dirA, err)
		return false
	}
	dirB, err = filepath.Abs(dirB)
	if err != nil {
		loggerE.Fatalf("filepath.Abs(%v)", dirB, err)
		return false
	}

	return strings.HasPrefix(dirA, dirB)
}
