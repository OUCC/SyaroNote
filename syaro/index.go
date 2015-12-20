package main

import (
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

				// register name
				name := removeExt(fi.Name())
				wpath := filepath.Join(wdir, fi.Name())
				elem, present := nIdx[name]
				if present {
					nIdx[name] = append(elem, wpath)
				} else {
					nIdx[name] = []string{wpath}
				}

				// TODO alias

				if setting.search {
					b, err := ioutil.ReadFile(filepath.Join(setting.wikiRoot, wpath))
					if err == nil {
						bleveIdx.Index(wpath, string(b))
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
