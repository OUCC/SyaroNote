package wikiio

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/util"

	"github.com/libgit2/git2go"

	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

// WikiPathList returns slice of each WikiFile in wikipath
// (slice doesn't include urlPrefix)
func (f *WikiFile) WikiPathList() []*WikiFile {
	Log.Debug("building...")
	s := strings.Split(util.RemoveExt(f.WikiPath()), "/")
	if s[0] == "" {
		s = s[1:]
	}

	ret := make([]*WikiFile, len(s))
	for i := 0; i < len(ret); i++ {
		path := "/" + strings.Join(s[:i+1], "/")
		Log.Debug("load %s", path)
		//		p, err := LoadPage(path)
		wfile, err := Load(path)
		if err != nil {
			Log.Debug("error in wikiio.Load(path): %s", err)
		}
		ret[i] = wfile
	}
	Log.Debug("finish")
	return ret
}

// WIKIROOT/a/b/c.md
func (f *WikiFile) FilePath() string {
	return filepath.Join(setting.WikiRoot, f.wikiPath)
}

// URLPREFIX/a/b/c.md
func (f *WikiFile) URLPath() template.URL {
	path := filepath.Join(setting.UrlPrefix, f.wikiPath)

	// url escape and revert %2F -> /
	return template.URL(strings.Replace(url.QueryEscape(path), "%2F", "/", -1))
}

func (f *WikiFile) IsDir() bool { return f.fileInfo.IsDir() }

func (f *WikiFile) DirMainPage() *WikiFile {
	if !f.IsDir() {
		return nil
	}

	var name string
	if f.WikiPath() == "/" {
		name = "Home"
	} else {
		name = f.Name()
	}

	for _, file := range f.files {
		if file.NameWithoutExt() == name {
			return file
		}
	}

	// not found
	return nil
}

func (f *WikiFile) IsMarkdown() bool { return util.IsMarkdown(f.wikiPath) }

func (f *WikiFile) Files() []*WikiFile { return f.files }

func (f *WikiFile) ParentDir() *WikiFile { return f.parentDir }

func (f *WikiFile) Raw() []byte {
	if f.IsDir() {
		return nil
	}

	b, err := ioutil.ReadFile(f.FilePath())
	if err != nil {
		Log.Fatal(err)
	}

	return b
}

func (f *WikiFile) Save(b []byte) error {
	if err := ioutil.WriteFile(f.FilePath(), b, 0644); err != nil {
		return err
	}

	// git commit
	if setting.GitMode {
		// get signature
		// TODO custom signature
		sig := getDefaultSignature()

		commit, err := commitChange(
			func(idx *git.Index) error {
				if err := idx.AddByPath(f.wikiPath[1:]); err != nil {
					return err
				}
				return nil
			},
			sig,
			"Updated "+filepath.Base(f.wikiPath))
		if err != nil {
			Log.Error("Git error: %s", err)
			return nil // dont send git error to client
		}
		defer commit.Free()
		logCommit(commit)
	}
	return nil
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

	// git commit
	if setting.GitMode {
		// get signature
		sig := getDefaultSignature()

		commit, err := commitChange(
			func(idx *git.Index) error {
				if err := idx.RemoveByPath(f.wikiPath[1:]); err != nil {
					return err
				}
				return nil
			},
			sig,
			"Removed "+filepath.Base(f.wikiPath))
		if err != nil {
			Log.Error("Git error: %s", err)
			return nil // dont send git error to client
		}
		defer commit.Free()
		logCommit(commit)
	}

	return nil
}
