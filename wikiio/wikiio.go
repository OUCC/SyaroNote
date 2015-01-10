package wikiio

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/util"

	"github.com/libgit2/git2go"
	"gopkg.in/fsnotify.v1"

	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	WikiRoot    *WikiFile
	searchIndex map[string][]*WikiFile
	watcher     *fsnotify.Watcher
	repo        *git.Repository
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
	const HIDDEN_DIR = "/."

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
					BuildIndex()
					Log.Info("File index refreshed")

				case event.Op&fsnotify.Remove != 0:
					Log.Info("File removed (%s)", event.Name)
					BuildIndex()
					Log.Info("File index refreshed")
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

		if info.IsDir() && !strings.Contains(path, HIDDEN_DIR) {
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
func BuildIndex() {
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
}

func Load(wpath string) (*WikiFile, error) {
	Log.Debug("wikiio.Load(%s)", wpath)

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
	Log.Debug("wikiio.Search(%s)", name)
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

	// git commit
	if repository != nil {
		status, err := repository.StatusFile(wpath[1:])
		if err != nil {
			Log.Error("Git error: %s", err)
			return nil
		}
		Log.Debug("status: %d", status)
		if status != git.StatusWtNew {
			Log.Error("Git error: Invalid status")
			return nil
		}
		oid, err := repository.CreateCommit("HEAD", nil, nil, "Created "+
			filepath.Base(wpath), nil)
		if err != nil {
			Log.Error("Git error: %s", err)
			return nil
		}
		commit, err := repository.LookupCommit(oid)
		if err != nil {
			Log.Error("Git error: %s", err)
			return nil
		}
		Log.Notice("Git commit created")
		Log.Info("Message: %s", commit.Message())
		Log.Info("Author: %s <%s>", commit.Author().Name, commit.Author().Email)
		Log.Info("Committer: %s <%s>", commit.Committer().Name, commit.Committer().Email)
	}

	return nil
}

func Rename(oldpath string, newpath string) error {
	Log.Debug("wikiio.Rename(%s, %s)", oldpath, newpath)

	f, err := Load(oldpath)
	if err != nil {
		return err
	}

	if !util.IsMarkdown(newpath) {
		newpath += ".md"
	}

	path := filepath.Join(setting.WikiRoot, newpath)
	os.MkdirAll(filepath.Dir(path), 0755)
	err = os.Rename(f.FilePath(), path)
	if err != nil {
		Log.Debug("can't rename: %s", err)
		return err
	}

	return nil
}
