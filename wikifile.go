package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	NEW_MD = "_New.md"

	WIKIFILE_FOLDER = 1 + iota
	WIKIFILE_MARKDOWN
	WIKIFILE_OTHER
)

type WikiFile struct {
	os.FileInfo

	WikiPath string
	fileType int
}

func loadFile(wpath string) (WikiFile, error) {
	if wpath == "" {
		wpath = string(filepath.Separator)
	}
	wpath = filepath.Clean(wpath)

	// TODO security check

	fpath := filepath.Join(setting.wikiRoot, wpath)
	fi, err := os.Stat(fpath)
	if err != nil {
		return WikiFile{}, err
	}

	ft := 0
	if fi.IsDir() {
		ft = WIKIFILE_FOLDER
	} else if isMarkdown(wpath) {
		ft = WIKIFILE_MARKDOWN
	} else {
		ft = WIKIFILE_OTHER
	}

	log.Debug("wpath: %s, ft: %d", wpath, ft)

	return WikiFile{
		FileInfo: fi,
		WikiPath: wpath,
		fileType: ft,
	}, nil
}

func createFile(wpath string) (WikiFile, error) {
	// TODO security check
	fpath := filepath.Join(setting.wikiRoot, wpath)

	// check if file is already exists
	_, err := os.Stat(fpath)
	if err == nil { // already exists
		return WikiFile{}, os.ErrExist
	}

	os.MkdirAll(filepath.Dir(fpath), 0755)
	err = ioutil.WriteFile(fpath, []byte("\n"), 0644)
	if err != nil {
		log.Debug(err.Error())
		return WikiFile{}, err
	}

	// if the new page template is exists
	tmplPath := filepath.Join(setting.wikiRoot, NEW_MD)
	if src, err := os.Open(tmplPath); err == nil {
		log.Debug("%s found. copy", tmplPath)
		// use template
		dst, _ := os.Open(fpath)
		io.Copy(src, dst)
	}
	return loadFile(wpath)
}

func (wf WikiFile) NameWithoutExt() string {
	return removeExt(wf.Name())
}

// URLPREFIX/a/b/c.md
func (wf WikiFile) URL() template.URL {
	wpath := filepath.Join(setting.urlPrefix, wf.WikiPath)
	s := queryEscapeWikiPath(wpath)
	return template.URL(s)
}

func (wf WikiFile) path() string {
	return filepath.Join(setting.wikiRoot, wf.WikiPath)
}

func (wf WikiFile) files() []WikiFile {
	if !wf.IsDir() {
		return nil
	}

	fis, _ := ioutil.ReadDir(wf.path())
	ret := make([]WikiFile, len(fis))
	for i, fi := range fis {
		wf_ := WikiFile{
			FileInfo: fi,
			WikiPath: filepath.Join(wf.WikiPath, fi.Name()),
		}
		switch {
		case fi.IsDir():
			wf_.fileType = WIKIFILE_FOLDER
			ret[i] = wf_
		case isMarkdown(fi.Name()):
			wf_.fileType = WIKIFILE_MARKDOWN
			ret[i] = wf_
		default:
			wf_.fileType = WIKIFILE_OTHER
			ret[i] = wf_
		}
	}
	return ret
}

func (wf WikiFile) parent() (WikiFile, bool) {
	if wf.WikiPath == string(filepath.Separator) {
		return WikiFile{}, false
	}

	dir := filepath.Dir(wf.WikiPath)
	fi, _ := os.Stat(filepath.Join(setting.wikiRoot, dir))
	return WikiFile{
		FileInfo: fi,
		WikiPath: dir,
		fileType: WIKIFILE_FOLDER,
	}, true
}

func (wf WikiFile) read() ([]byte, error) {
	return ioutil.ReadFile(wf.path())
}

func (wf WikiFile) save(b []byte) error {
	return ioutil.WriteFile(wf.path(), b, 0644)
}

func (wf WikiFile) remove() error {
	return os.RemoveAll(wf.path())
}

func (wf WikiFile) rename(newpath string) error {
	path := filepath.Join(setting.wikiRoot, newpath)
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.Rename(wf.path(), path)
}
