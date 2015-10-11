package main

import (
	pb "github.com/OUCC/syaro/gitservice"
	"github.com/OUCC/syaro/markdown"

	"golang.org/x/net/context"

	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func getPage(w http.ResponseWriter, r *http.Request) {
	wpath := r.URL.Query().Get("wpath")

	log.Info("Loading page (%s)...", wpath)
	wf, err := loadFile(wpath)
	if os.IsNotExist(err) {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := wf.read()
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(markdown.Convert(b, filepath.Dir(wpath)))
	log.Info("OK")
}

func createPage(w http.ResponseWriter, r *http.Request) {
	wpath := r.URL.Query().Get("wpath")

	if !isMarkdown(wpath) {
		wpath += ".md"
	}
	log.Info("Creating new page (%s)...", wpath)

	_, err := createFile(wpath)
	if os.IsExist(err) {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusFound)
		return
	} else if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send success response
	http.Error(w, "OK", http.StatusCreated)
	log.Info("OK")

	if setting.gitMode {
		err := gitCommit(func(client pb.GitClient) (*pb.CommitResponse, error) {
			return client.Save(context.Background(), &pb.SaveRequest{
				Path: wpath,
				Msg:  "Created " + wpath,
			})
		})
		if err != nil {
			log.Error(err.Error())
		}
	}
	// TODO postSave
}

func updatePage(w http.ResponseWriter, r *http.Request) {
	wpath := r.URL.Query().Get("wpath")
	msg := r.URL.Query().Get("message")
	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")

	log.Info("Saving (%s)...", wpath)

	wf, err := loadFile(wpath)
	if os.IsNotExist(err) {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err = wf.save(b); err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, "OK", http.StatusOK)
	log.Info("OK")

	if setting.gitMode {
		if msg == "" {
			msg = "Updated " + wpath
		}
		err := gitCommit(func(client pb.GitClient) (*pb.CommitResponse, error) {
			return client.Save(context.Background(), &pb.SaveRequest{
				Path:  wpath,
				Msg:   msg,
				Name:  name,
				Email: email,
			})
		})
		if err != nil {
			log.Error(err.Error())
		}
	}
	// TODO postSave
}

func renameFile(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Query().Get("src")
	dst := r.URL.Query().Get("dst")

	wf, err := loadFile(src)
	if os.IsNotExist(err) {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if wf.fileType == WIKIFILE_MARKDOWN && !isMarkdown(dst) {
		dst += ".md"
	}

	log.Info("Rename page (%s -> %s)...", src, dst)

	if err := wf.rename(dst); err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send success response
	http.Error(w, "OK", http.StatusOK)
	log.Info("OK")

	if setting.gitMode {
		gitCommit(func(client pb.GitClient) (*pb.CommitResponse, error) {
			return client.Rename(context.Background(), &pb.RenameRequest{
				Src: src,
				Dst: dst,
				Msg: fmt.Sprintf("Renamed %s -> %s", src, dst),
			})
		})
	}
	// TODO postSave
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	wpath := r.URL.Query().Get("wpath")
	log.Info("Deleting page (%s)...", wpath)

	wf, err := loadFile(wpath)
	if os.IsNotExist(err) {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = wf.remove(); err != nil {
		log.Error("Delete file error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send success response
	http.Error(w, "OK", http.StatusOK)
	log.Info("Deleted")

	if setting.gitMode {
		err := gitCommit(func(client pb.GitClient) (*pb.CommitResponse, error) {
			return client.Remove(context.Background(), &pb.RemoveRequest{
				Path: wpath,
				Msg:  "Removed " + wpath,
			})
		})
		if err != nil {
			log.Error(err.Error())
		}
	}
	// TODO post delete
}

func searchPage(w http.ResponseWriter, r *http.Request) {
}

/*
func previewPage(w http.ResponseWriter, r *http.Request) {
	wpath := r.URL.Query().Get("wpath")
	log.Info("Rendering preview...")

	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	// return generated html
	w.Write(convertMd(b, filepath.Dir(wpath)))
}
*/

func uploadFile(w http.ResponseWriter, r *http.Request) {
}
