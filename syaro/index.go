package main

import (
	"github.com/OUCC/SyaroNote/syaro/markdown"

	"github.com/blevesearch/bleve"

	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	BLEVE_PATH = ".syaronote.bleve"
)

var (
	// [page name, alias] -> []WikiPath
	nameIdx = make(map[string][]string)

	// [tag name] -> []WikiPath
	tagIdx map[string][]string // TODO

	// mutex for index
	idxMtx = new(sync.Mutex)

	bleveIdx bleve.Index

	refresh = make(chan string)
)

// must be called after setting.wikiRoot is set
func idxBuilder() {
	if setting.search {
		blevePath := filepath.Join(setting.wikiRoot, BLEVE_PATH)
		os.RemoveAll(blevePath)
		mapping := bleve.NewIndexMapping()
		var err error
		bleveIdx, err = bleve.New(blevePath, mapping)
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	// builder loop
	for {
		wpath := <-refresh
		log.Info("Rebuilding index... (wpath: %s)", wpath)

		nIdx := make(map[string][]string)
		// tIdx := make(map[string][]string)

		var walkfunc func(string)
		walkfunc = func(wdir string) {
			infos, _ := ioutil.ReadDir(filepath.Join(setting.wikiRoot, wdir))

			for _, fi := range infos {
				// skip hidden files and backup files
				if strings.HasPrefix(fi.Name(), ".") {
					continue
				}
				wpath := filepath.Join(wdir, fi.Name())

				// register name
				if fi.IsDir() || isMarkdown(fi.Name()) {
					name := removeExt(fi.Name())
					elem, present := nIdx[name]
					if present {
						nIdx[name] = append(elem, wpath)
					} else {
						nIdx[name] = []string{wpath}
					}
				}

				// register alias
				if !fi.IsDir() && isMarkdown(fi.Name()) {
					b, err := ioutil.ReadFile(filepath.Join(setting.wikiRoot, wpath))
					if err != nil {
						log.Error("Failed to read file %s: %v", wpath, err)
						continue
					}
					for _, alias := range strings.Split(markdown.Meta(b)["alias"], ",") {
						alias = strings.TrimSpace(alias)
						if alias == "" {
							continue
						}
						elem, present := nIdx[alias]
						if present {
							nIdx[alias] = append(elem, wpath)
						} else {
							nIdx[alias] = []string{wpath}
						}
					}

					if setting.search {
						bleveIdx.Delete(wpath)
						bleveIdx.Index(wpath, b)
					}
				}

				if fi.IsDir() {
					walkfunc(wpath)
				}
			}
		}
		walkfunc(wpath)

		idxMtx.Lock()
		nameIdx = nIdx
		// tagIdx = tIdx
		idxMtx.Unlock()

		log.Info("Index building end")
	}
}
