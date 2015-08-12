package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	wikiRoot  *WikiFile
	wikiIndex map[string][]*WikiFile

	// if true, BuildIndex is called
	refreshRequired = true

	ErrNotExist = errors.New("file not exist")
	ErrNotFound = errors.New("file not found")
)

// must be called after setting.wikiRoot is set
func buildIndex() {
	log.Debug("Index building start")

	info, err := os.Stat(setting.wikiRoot)
	if err != nil {
		log.Fatal(err)
	}

	wikiRoot = &WikiFile{
		parentDir: nil,
		wikiPath:  "/",
		fileInfo:  info,
	}
	wikiIndex = make(map[string][]*WikiFile)

	// anonymous recursive function
	var walkfunc func(*WikiFile)
	walkfunc = func(dir *WikiFile) {
		infos, _ := ioutil.ReadDir(filepath.Join(setting.wikiRoot, dir.WikiPath()))

		dir.files = make([]*WikiFile, 0, len(infos))
		for _, info := range infos {
			// skip hidden files and backup files
			if info.Name()[:1] == "." || strings.HasSuffix(info.Name(), BACKUP_SUFFIX) {
				continue
			}

			file := &WikiFile{
				parentDir: dir,
				wikiPath:  filepath.Join(dir.WikiPath(), info.Name()),
				fileInfo:  info,
			}
			dir.files = append(dir.files, file)

			// register to wikiIndex
			elem, present := wikiIndex[file.Name()]
			if present {
				wikiIndex[file.Name()] = append(elem, file)
			} else {
				wikiIndex[file.Name()] = []*WikiFile{file}
			}

			elem, present = wikiIndex[file.NameWithoutExt()]
			if present {
				wikiIndex[file.NameWithoutExt()] = append(elem, file)
			} else {
				wikiIndex[file.NameWithoutExt()] = []*WikiFile{file}
			}

			if info.IsDir() {
				walkfunc(file)
			}
		}
	}
	walkfunc(wikiRoot)

	log.Debug("Index building end")
	log.Info("File index refreshed")

	refreshRequired = false
}

func searchFile(name string) ([]*WikiFile, error) {
	log.Debug("name: %s", name)

	if refreshRequired {
		buildIndex()
	}

	files, present := wikiIndex[name]
	if !present {
		log.Debug("not found")
		return nil, ErrNotFound
	}

	// for debug output
	found := make([]string, len(files))
	for i := 0; i < len(found); i++ {
		found[i] = files[i].WikiPath()
	}
	log.Debug("found %v", found)

	return files, nil
}
