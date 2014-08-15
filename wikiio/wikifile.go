package wikiio

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/util"

	"io/ioutil"
	"os"
	"path/filepath"
)

type WikiFile struct {
	fileInfo  os.FileInfo
	files     []*WikiFile
	parentDir *WikiFile
	wikiPath  string
}

// Base name of file (with ext)
func (f *WikiFile) Name() string {
	return filepath.Base(f.wikiPath)
}

func (f *WikiFile) NameWithoutExt() string {
	return util.RemoveExt(f.Name())
}

func (f *WikiFile) WikiPath() string { return f.wikiPath }

// WIKIROOT/a/b/c.md
func (f *WikiFile) FilePath() string {
	return filepath.Join(setting.WikiRoot, f.wikiPath)
}

// URLPREFIX/a/b/c.md
func (f *WikiFile) URLPath() string {
	return filepath.Join(setting.UrlPrefix, f.wikiPath)
}

func (f *WikiFile) IsDir() bool { return f.fileInfo.IsDir() }

func (f *WikiFile) IsMarkdown() bool { return util.IsMarkdown(f.wikiPath) }

func (f *WikiFile) Files() []*WikiFile { return f.files }

func (f *WikiFile) ParentDir() *WikiFile { return f.parentDir }

func (f *WikiFile) Raw() []byte {
	if f.IsDir() {
		return nil
	}

	b, err := ioutil.ReadFile(f.FilePath())
	if err != nil {
		LoggerE.Fatalln(err)
	}

	return b
}

func (f *WikiFile) Remove() error {
	if err := os.Remove(f.FilePath()); err != nil {
		return err
	}

	// // remove f from parentDir.Files
	// tmp := make([]*WikiFile, len(f.parentDir.files)-1)
	// for _, file := range f.parentDir.files {
	// 	if file.Name() != f.Name() {
	// 		tmp := append(tmp, file)
	// 	}
	// }
	// f.parentDir.files = tmp

	// // remove f from searchIndex

	// FIXME
	// update index
	BuildIndex()
	return nil
}
