package main

import (
	pb "github.com/OUCC/syaro/gitservice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func loadFile(wpath string) (*WikiFile, error) {
	log.Debug("wpath: %s", wpath)

	if refreshRequired {
		buildIndex()
	}

	// wiki root
	if wpath == "/" || wpath == "." || wpath == "" {
		return wikiRoot, nil
	}

	sl := strings.Split(wpath, "/")
	ret := wikiRoot
	for _, s := range sl {
		if s == "" {
			continue
		}

		tmp := ret
		for _, f := range ret.Files() {
			if f.Name() == s || removeExt(f.Name()) == s {
				ret = f
				break
			}
		}
		// not found
		if ret == tmp {
			log.Debug("wikiio.Load: not exist")
			return nil, ErrNotExist
		}
	}

	return ret, nil
}

func createFile(wpath string) error {
	log.Debug("wpath: %s", wpath)

	initialText := removeExt(filepath.Base(wpath)) + "\n====\n"

	// check if file is already exists
	file, _ := loadFile(wpath)
	if file != nil {
		// if exists, return error
		return os.ErrExist
	}

	if !isMarkdown(wpath) {
		wpath += ".md"
	}

	path := filepath.Join(setting.wikiRoot, wpath)
	os.MkdirAll(filepath.Dir(path), 0755)
	err := ioutil.WriteFile(path, []byte(initialText), 0644)
	if err != nil {
		log.Debug(err.Error())
		return err
	}

	refreshRequired = true

	return gitCommit(func(client pb.GitClient) (*pb.CommitResponse, error) {
		return client.Save(context.Background(), &pb.SaveRequest{
			Path: wpath,
			Msg:  "Created " + wpath,
		})
	})
}

func renameFile(oldpath string, newpath string) error {
	log.Debug("oldpath: %s, newpath: %s", oldpath, newpath)

	f, err := loadFile(oldpath)
	if err != nil {
		return err
	}

	if !f.IsDir() && f.IsMarkdown() && !isMarkdown(newpath) {
		newpath += ".md"
	}

	path := filepath.Join(setting.wikiRoot, newpath)
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.Rename(f.FilePath(), path); err != nil {
		log.Debug("can't rename: %s", err)
		return err
	}

	refreshRequired = true

	return gitCommit(func(client pb.GitClient) (*pb.CommitResponse, error) {
		return client.Rename(context.Background(), &pb.RenameRequest{
			Src: oldpath,
			Dst: newpath,
			Msg: fmt.Sprintf("Renamed %s -> %s", oldpath, newpath),
		})
	})
}

func gitCommit(commitFunc func(pb.GitClient) (*pb.CommitResponse, error)) error {
	if !setting.gitMode {
		return nil
	}

	conn, err := grpc.Dial("127.0.0.1:" + strconv.Itoa(setting.port+1))
	if err != nil {
		log.Debug("Dial error: %s", err)
		return err
	}
	defer conn.Close()

	client := pb.NewGitClient(conn)
	res, err := commitFunc(client)
	if err != nil {
		log.Debug("Git error: %s", err)
		return err
	}
	log.Debug("commit id: %s, message: %s", res.Msg, res.Msg)
	return nil
}
