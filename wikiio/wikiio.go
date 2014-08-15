package wikiio

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/util"

	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	WikiRoot    *WikiFile
	searchIndex = make(map[string][]*WikiFile)
)

var (
	ErrNotExist = errors.New("file not exist")
	ErrNotFound = errors.New("file not found")
)

// must be called after setting.WikiRoot is set
func BuildIndex() {
	LoggerV.Println("wikiio.BuildIndex(): Start")

	info, err := os.Stat(setting.WikiRoot)
	if err != nil {
		LoggerE.Fatalln("Error:", err)
	}

	WikiRoot = &WikiFile{
		parentDir: nil,
		wikiPath:  "/",
		fileInfo:  info,
	}

	walkfunc(WikiRoot)

	LoggerV.Println("wikiio.BuildIndex(): End")
}

// func for rescursive
func walkfunc(dir *WikiFile) {
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

func Load(wpath string) (*WikiFile, error) {
	LoggerV.Printf("wikiio.Load(%s)", wpath)

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
			LoggerV.Printf("wikiio.Load: not exist")
			return nil, ErrNotExist
		}
	}

	return ret, nil
}

func Search(name string) ([]*WikiFile, error) {
	LoggerV.Printf("wikiio.Search(%s)", name)
	files, present := searchIndex[name]
	if !present {
		LoggerV.Println("wikiio.Search: not found")
		return nil, ErrNotFound
	}

	// for debug output
	found := make([]string, len(files))
	for i := 0; i < len(found); i++ {
		found[i] = files[i].WikiPath()
	}
	LoggerV.Println("wikiio.Search: found", found)

	return files, nil
}

func Create(wpath string) *WikiFile {
	_, err := os.Create(filepath.Join(setting.WikiRoot, wpath))
	if err != nil {
		LoggerE.Fatalln("Error (os.Create):", err)
	}

	// FIXME
	BuildIndex()
	return nil
}
