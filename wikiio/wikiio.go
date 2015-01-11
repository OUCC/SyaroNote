package wikiio

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/util"

	"github.com/libgit2/git2go"
	"gopkg.in/fsnotify.v1"

	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	WikiRoot    *WikiFile
	searchIndex map[string][]*WikiFile

	// git repository
	repo *git.Repository

	// file system watcher
	watcher *fsnotify.Watcher

	// if true, BuildIndex is called
	refreshRequired = true
)

var (
	ErrNotExist     = errors.New("file not exist")
	ErrNotFound     = errors.New("file not found")
	ErrRepoNotReady = errors.New("repository contains uncommited changes")
)

func OpenRepository() error {
	var err error
	repo, err = git.OpenRepository(setting.WikiRoot)
	if err != nil {
		return err
	}

	// check if repository contains uncommited changes
	opt := new(git.StatusOptions)
	opt.Flags = git.StatusOptIncludeUntracked
	opt.Show = git.StatusShowIndexAndWorkdir
	statuses, err := repo.StatusList(opt)
	if err != nil {
		return err
	} else {
		defer statuses.Free()
	}
	if c, _ := statuses.EntryCount(); c != 0 {
		return ErrRepoNotReady
	}

	return nil
}

func InitWatcher() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		Log.Fatal(err)
	}

	// event loop for watcher
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				Log.Debug("%s", event)
				switch {
				case event.Op&fsnotify.Create != 0:
					Log.Info("New file Created (%s)", event.Name)
					refreshRequired = true

				case event.Op&fsnotify.Remove != 0:
					Log.Info("File removed (%s)", event.Name)
					refreshRequired = true
				}

			case err := <-watcher.Errors:
				Log.Fatal(err)
			}
		}
	}()

	filepath.Walk(setting.WikiRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			Log.Error(err.Error())
		}

		// dont add hidden dir (ex. .git)
		if info.IsDir() && !strings.Contains(path, "/.") && !strings.HasPrefix(path, ".") {
			watcher.Add(path)
			Log.Debug("%s added to watcher", path)
		}

		return nil
	})
}

func CloseWatcher() {
	watcher.Close()
}

// must be called after setting.WikiRoot is set
func buildIndex() {
	Log.Debug("Index building start")

	info, err := os.Stat(setting.WikiRoot)
	if err != nil {
		Log.Fatal(err)
	}

	WikiRoot = &WikiFile{
		parentDir: nil,
		wikiPath:  "/",
		fileInfo:  info,
	}
	searchIndex = make(map[string][]*WikiFile)

	// anonymous recursive function
	var walkfunc func(*WikiFile)
	walkfunc = func(dir *WikiFile) {
		infos, _ := ioutil.ReadDir(filepath.Join(setting.WikiRoot, dir.WikiPath()))

		dir.files = make([]*WikiFile, 0, len(infos))
		for _, info := range infos {
			// skip hidden file
			if info.Name()[:1] == "." {
				continue
			}

			file := &WikiFile{
				parentDir: dir,
				wikiPath:  filepath.Join(dir.WikiPath(), info.Name()),
				fileInfo:  info,
			}
			dir.files = append(dir.files, file)

			// register to searchIndex
			elem, present := searchIndex[file.Name()]
			if present {
				searchIndex[file.Name()] = append(elem, file)
			} else {
				searchIndex[file.Name()] = []*WikiFile{file}
			}

			elem, present = searchIndex[file.NameWithoutExt()]
			if present {
				searchIndex[file.NameWithoutExt()] = append(elem, file)
			} else {
				searchIndex[file.NameWithoutExt()] = []*WikiFile{file}
			}

			if info.IsDir() {
				walkfunc(file)
			}
		}
	}
	walkfunc(WikiRoot)

	Log.Debug("Index building end")
	Log.Info("File index refreshed")

	refreshRequired = false
}

func Load(wpath string) (*WikiFile, error) {
	Log.Debug("wpath: %s", wpath)

	if refreshRequired {
		buildIndex()
	}

	// wiki root
	if wpath == "/" || wpath == "." || wpath == "" {
		return WikiRoot, nil
	}

	sl := strings.Split(wpath, "/")
	ret := WikiRoot
	for _, s := range sl {
		if s == "" {
			continue
		}

		tmp := ret
		for _, f := range ret.Files() {
			if f.Name() == s || util.RemoveExt(f.Name()) == s {
				ret = f
				break
			}
		}
		// not found
		if ret == tmp {
			Log.Debug("wikiio.Load: not exist")
			return nil, ErrNotExist
		}
	}

	return ret, nil
}

func Search(name string) ([]*WikiFile, error) {
	Log.Debug("name: %s", name)

	if refreshRequired {
		buildIndex()
	}

	files, present := searchIndex[name]
	if !present {
		Log.Debug("not found")
		return nil, ErrNotFound
	}

	// for debug output
	found := make([]string, len(files))
	for i := 0; i < len(found); i++ {
		found[i] = files[i].WikiPath()
	}
	Log.Debug("found %v", found)

	return files, nil
}

func Create(wpath string) error {
	Log.Debug("wpath: %s", wpath)

	initialText := util.RemoveExt(filepath.Base(wpath)) + "\n====\n"

	// check if file is already exists
	file, _ := Load(wpath)
	if file != nil {
		// if exists, return error
		return os.ErrExist
	}

	if !util.IsMarkdown(wpath) {
		wpath += ".md"
	}

	path := filepath.Join(setting.WikiRoot, wpath)
	os.MkdirAll(filepath.Dir(path), 0755)
	err := ioutil.WriteFile(path, []byte(initialText), 0644)
	if err != nil {
		Log.Debug(err.Error())
		return err
	}

	refreshRequired = true

	// git commit
	if setting.GitMode {
		// get signature
		sig := getDefaultSignature()

		commit, err := commitChange(
			func(idx *git.Index) error {
				if err := idx.AddByPath(wpath[1:]); err != nil {
					return err
				}
				return nil
			},
			sig,
			"Created "+filepath.Base(wpath))
		if err != nil {
			Log.Error("Git error: %s", err)
			return nil // dont send git error to client
		}
		defer commit.Free()
		logCommit(commit)
	}

	return nil
}

func Rename(oldpath string, newpath string) error {
	Log.Debug("oldpath: %s, newpath: %s", oldpath, newpath)

	f, err := Load(oldpath)
	if err != nil {
		return err
	}

	if !f.IsDir() && f.IsMarkdown() && !util.IsMarkdown(newpath) {
		newpath += ".md"
	}

	path := filepath.Join(setting.WikiRoot, newpath)
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.Rename(f.FilePath(), path); err != nil {
		Log.Debug("can't rename: %s", err)
		return err
	}

	refreshRequired = true

	// git commit
	if setting.GitMode {
		// get signature
		sig := getDefaultSignature()

		commit, err := commitChange(
			func(idx *git.Index) error {
				if err := idx.RemoveByPath(oldpath[1:]); err != nil {
					return err
				}
				if err := idx.AddByPath(newpath[1:]); err != nil {
					return err
				}
				return nil
			},
			sig,
			fmt.Sprintf("Renamed %s\n\n%s -> %s", filepath.Base(oldpath), oldpath, newpath))

		if err != nil {
			Log.Error("Git error: %s", err)
			return nil // dont send git error to client
		}
		defer commit.Free()
		logCommit(commit)
	}

	return nil
}
