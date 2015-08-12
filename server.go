package main

import (
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"io/ioutil"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type wikiHandler func(wpath string, w http.ResponseWriter, r *http.Request)

func startServer() {
	// fix url prefix
	if setting.urlPrefix != "" {
		setting.urlPrefix = filepath.Clean("/" + setting.urlPrefix)
	}

	// set handlers
	mux := web.New()
	mux.Use(middleware.RequestID)
	if setting.verbose {
		mux.Use(middleware.Logger)
	}
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.AutomaticOptions)

	// files under SYARO_DIR/public
	rootDir := http.Dir(filepath.Join(setting.syaroDir, PUBLIC_DIR))
	fileServer := http.StripPrefix(setting.urlPrefix, http.FileServer(rootDir))
	mux.Get(setting.urlPrefix+"/css/*", fileServer)
	mux.Get(setting.urlPrefix+"/fonts/*", fileServer)
	mux.Get(setting.urlPrefix+"/ico/*", fileServer)
	mux.Get(setting.urlPrefix+"/js/*", fileServer)
	mux.Get(setting.urlPrefix+"/lib/*", fileServer)

	mux.Get(setting.urlPrefix+"/error/:code",
		func(c web.C, w http.ResponseWriter, r *http.Request) {
			i, _ := strconv.Atoi(c.URLParams["code"])
			if i == 0 { // invalid request
				i = 400
			}
			errorHandler(w, i, r.URL.Query().Get("data"))
		})
	mux.Get(setting.urlPrefix+"/*",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Query().Get("view") {
			case "":
				handlerConverter(viewPage)(w, r)
			case "editor":
				handlerConverter(editorView)(w, r)
			case "history":
				handlerConverter(historyView)(w, r)
			default:
				data := r.URL.Query().Get("view")
				log.Error("invalid URL query (view: %s)", data)
				errorHandler(w, http.StatusBadRequest, data)
			}
		})
	mux.Post(setting.urlPrefix+"/*", handlerConverter(createPage))
	mux.Put(setting.urlPrefix+"/*",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Query().Get("action") {
			case "":
				handlerConverter(updatePage)(w, r)
			case "rename":
				handlerConverter(renamePage)(w, r)
			default:
				data := r.URL.Query().Get("action")
				log.Error("invalid URL query (action: %s)", data)
				errorHandler(w, http.StatusBadRequest, data)
			}
		})
	mux.Delete(setting.urlPrefix+"/*", handlerConverter(deletePage))

	log.Notice("Server started. Waiting connection localhost:%d%s\n",
		setting.port, setting.urlPrefix)

	// listen port
	l, err := net.Listen("tcp", ":"+strconv.Itoa(setting.port))
	if err != nil {
		log.Fatal(err)
	}

	if setting.fcgi {
		if err := fcgi.Serve(l, mux); err != nil {
			log.Fatal(err)
		}
	} else {
		http.Handle("/", mux)
		if err = graceful.Serve(l, http.DefaultServeMux); err != nil {
			log.Fatal(err)
		}
		graceful.Wait()
	}
}

func handlerConverter(f wikiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// url unescape (+ -> <Space>)
		r.URL.Path = strings.Replace(r.URL.Path, "+", " ", -1)
		wpath := strings.TrimPrefix(r.URL.Path, setting.urlPrefix)

		log.Debug("WikiPath: %s, Query: %s, Fragment: %s", wpath,
			r.URL.RawQuery, r.URL.Fragment)
		f(wpath, w, r)
	}
}

func viewPage(wpath string, w http.ResponseWriter, r *http.Request) {
	v, err := LoadPage(wpath)
	switch err {
	case nil:
		// render html
		log.Info("Rendering page (%s)...", wpath)
		err = v.Render(w)
		if err != nil {
			log.Error("Rendering error!: %s", err)
			errorHandler(w, http.StatusInternalServerError, err.Error())
			return
		}
		log.Info("OK")

	case ErrIsNotMarkdown:
		log.Info("Sending file (%s)...", wpath)
		v, err := loadFile(wpath)
		if err != nil {
			log.Error(err.Error())
			errorHandler(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(v.Raw())
		log.Info("OK")

	default:
		log.Error("%s (%s)", err, wpath)
		errorHandler(w, http.StatusNotFound, wpath)
		return
	}
}

func createPage(wpath string, w http.ResponseWriter, r *http.Request) {
	log.Info("Creating new page (%s)...", wpath)

	switch err := createFile(wpath); err {
	case nil:
		// send success response
		errorHandler(w, http.StatusCreated, "")
		log.Info("OK")
	case os.ErrExist:
		log.Error(err.Error())
		errorHandler(w, http.StatusFound, "")
	default:
		log.Error(err.Error())
		errorHandler(w, http.StatusInternalServerError, err.Error())
	}
}

func updatePage(wpath string, w http.ResponseWriter, r *http.Request) {
	backup, _ := strconv.ParseBool(r.URL.Query().Get("backup"))
	message, _ := url.QueryUnescape(r.URL.Query().Get("message"))
	name, _ := url.QueryUnescape(r.URL.Query().Get("name"))
	email, _ := url.QueryUnescape(r.URL.Query().Get("email"))

	if backup {
		log.Info("Backing up (%s)...", wpath+".bac")
	} else {
		log.Info("Saving (%s)...", wpath)
	}

	f, err := loadFile(wpath)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, wpath+"not found", http.StatusNotFound)
		return
	}

	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if backup {
		err = f.SaveBackup(b)
	} else {
		err = f.Save(b, message, name, email)
	}
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("Rendering preview...")
	// return generated html
	w.Write(parseMarkdown(b, filepath.Dir(wpath)))
	log.Info("OK")
}

func renamePage(wpath string, w http.ResponseWriter, r *http.Request) {
	oldpath, err := url.QueryUnescape(r.URL.Query().Get("oldpath"))
	if err != nil {
		log.Error("Unescape error: %s", err)
		code := http.StatusBadRequest
		http.Error(w, http.StatusText(code), code)
		return
	}

	log.Info("Rename page (%s -> %s)...", oldpath, wpath)

	if err := renameFile(oldpath, wpath); err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send success response
	w.Write(nil)
	log.Info("OK")
}

func deletePage(wpath string, w http.ResponseWriter, r *http.Request) {
	log.Info("Deleting page (%s)...", wpath)

	f, err := loadFile(wpath)
	switch err {
	case nil:

	case os.ErrNotExist:
		log.Error(err.Error())
		code := http.StatusNotFound
		http.Error(w, http.StatusText(code), code)
		return

	default:
		log.Error("Load file error: %s", err)
		return
	}

	if err = f.Remove(); err != nil {
		log.Error("Delete file error: %s", err)
		http.Error(w, "cannot delete file", http.StatusInternalServerError)
		return
	}

	// send success response
	w.Write(nil)
	log.Info("Deleted")
}

func editorView(wpath string, w http.ResponseWriter, r *http.Request) {
	log.Info("Editor requested (%s)", wpath)

	// check main dir
	f, _ := loadFile(wpath)
	if f != nil {
		if f.DirMainPage() != nil {
			log.Info("Requested file is dir, let's redirect to main file")
			http.Redirect(w, r,
				string(f.DirMainPage().URLPath())+"?view=editor", http.StatusFound)
			return
		} else if f.IsDir() {
			log.Info("Requested file is dir, but main file dosen't exists.")
			errorHandler(w, http.StatusNotFound, wpath)
			return
		}
	}

	v, err := NewEditor(wpath)
	if err != nil {
		log.Error(err.Error())
		errorHandler(w, http.StatusNotFound, wpath)
		return
	}

	// render html
	log.Info("Rendering view...")
	err = v.Render(w)
	if err != nil {
		log.Error("Rendering error!: %s", err)
		errorHandler(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Info("OK")
}

func historyView(wpath string, w http.ResponseWriter, r *http.Request) {
	log.Info("History view requested")

	v, err := LoadHistoryPage(wpath)
	if err != nil {
		log.Error("%s (%s)", err, wpath)
		errorHandler(w, http.StatusNotFound, wpath)
		return
	}

	// render html
	log.Info("Rendering history page (%s)...", wpath)
	err = v.Render(w)
	if err != nil {
		log.Error("Rendering error!: %s", err)
		errorHandler(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Info("OK")
}

func errorHandler(w http.ResponseWriter, status int, data string) {
	log.Info("Rendering error view... (status: %d, data: %s)", status, data)

	w.WriteHeader(status)
	err := views.ExecuteTemplate(w, strconv.Itoa(status)+".html", data)
	if err != nil {
		// template not available
		http.Error(w, http.StatusText(status), status)
	}
	log.Info("OK")
}
