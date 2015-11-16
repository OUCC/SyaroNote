package main

import (
	pb "github.com/OUCC/syaro/gitservice"

	"golang.org/x/net/context"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

	w.Write(b)
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
		log.Error(err.Error())
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

func uploadFile(w http.ResponseWriter, r *http.Request) {
	wpath := r.URL.Query().Get("wpath")

	log.Info("Saving uploaded file(%s)...", wpath)

	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if _, err := saveFile(wpath, b); err != nil {
		log.Error(err.Error())
		if os.IsExist(err) {
			http.Error(w, err.Error(), http.StatusFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	http.Error(w, "OK", http.StatusCreated)
	log.Info("OK")

	if setting.gitMode {
		msg := "Added " + wpath
		err := gitCommit(func(client pb.GitClient) (*pb.CommitResponse, error) {
			return client.Save(context.Background(), &pb.SaveRequest{
				Path: wpath,
				Msg:  msg,
			})
		})
		if err != nil {
			log.Error(err.Error())
		}
	}
	// TODO postUpload
}

func getHistory(w http.ResponseWriter, r *http.Request) {
	wpath := r.URL.Query().Get("wpath")

	if !setting.gitMode {
		http.Error(w, "Git mode not enabled", http.StatusBadRequest)
		return
	}

	log.Info("loading git history... (%s)", wpath)
	changes := getChanges(wpath)

	// convert []*pb.Change to JSON
	v := make([]map[string]string, len(changes))
	for i, c := range changes {
		m := make(map[string]string)
		switch c.Op {
		case pb.Change_OpNone:
			m["op"] = "None"
		case pb.Change_OpAdd:
			m["op"] = "Add"
		case pb.Change_OpRename:
			m["op"] = "Rename"
		case pb.Change_OpUpdate:
			m["op"] = "Update"
		}
		m["name"] = c.Name
		m["email"] = c.Email
		m["msg"] = c.Msg

		v[i] = m
	}

	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
	log.Info("OK")
}
