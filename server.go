package main

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/wikiio"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"

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
	if setting.UrlPrefix != "" {
		setting.UrlPrefix = filepath.Clean("/" + setting.UrlPrefix)
	}

	// set handlers
	// files under SYARO_DIR/public
	rootDir := http.Dir(filepath.Join(setting.SyaroDir, PUBLIC_DIR))
	fileServer := http.StripPrefix(setting.UrlPrefix, http.FileServer(rootDir))
	goji.Get(setting.UrlPrefix+"/css/*", fileServer)
	goji.Get(setting.UrlPrefix+"/fonts/*", fileServer)
	goji.Get(setting.UrlPrefix+"/ico/*", fileServer)
	goji.Get(setting.UrlPrefix+"/js/*", fileServer)
	goji.Get(setting.UrlPrefix+"/lib/*", fileServer)

	goji.Get(setting.UrlPrefix+"/error/:code",
		func(c web.C, w http.ResponseWriter, r *http.Request) {
			i, _ := strconv.Atoi(c.URLParams["code"])
			if i == 0 { // invalid request
				i = 400
			}
			errorHandler(w, i, r.URL.Query().Get("data"))
		})
	goji.Get(setting.UrlPrefix+"/*",
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
				Log.Error("invalid URL query (view: %s)", data)
				errorHandler(w, http.StatusBadRequest, data)
			}
		})
	goji.Post(setting.UrlPrefix+"/*", handlerConverter(createPage))
	goji.Put(setting.UrlPrefix+"/*",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Query().Get("action") {
			case "":
				handlerConverter(updatePage)(w, r)
			case "rename":
				handlerConverter(renamePage)(w, r)
			default:
				data := r.URL.Query().Get("action")
				Log.Error("invalid URL query (action: %s)", data)
				errorHandler(w, http.StatusBadRequest, data)
			}
		})
	goji.Delete(setting.UrlPrefix+"/*", handlerConverter(deletePage))

	Log.Notice("Server started. Waiting connection localhost:%d%s\n",
		setting.Port, setting.UrlPrefix)

	// listen port
	l, err := net.Listen("tcp", ":"+strconv.Itoa(setting.Port))
	if err != nil {
		Log.Fatal(err)
	}

	if setting.FCGI {
		if err := fcgi.Serve(l, goji.DefaultMux); err != nil {
			Log.Fatal(err)
		}
	} else {
		http.Handle("/", goji.DefaultMux)
		if err = graceful.Serve(l, http.DefaultServeMux); err != nil {
			Log.Fatal(err)
		}
		graceful.Wait()
	}
}

func handlerConverter(f wikiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// url unescape (+ -> <Space>)
		r.URL.Path = strings.Replace(r.URL.Path, "+", " ", -1)
		wpath := strings.TrimPrefix(r.URL.Path, setting.UrlPrefix)

		Log.Debug("WikiPath: %s, Query: %s, Fragment: %s", wpath,
			r.URL.RawQuery, r.URL.Fragment)
		f(wpath, w, r)
	}
}

func viewPage(wpath string, w http.ResponseWriter, r *http.Request) {
	v, err := LoadPage(wpath)
	switch err {
	case nil:
		// render html
		Log.Info("Rendering page (%s)...", wpath)
		err = v.Render(w)
		if err != nil {
			Log.Error("Rendering error!: %s", err)
			errorHandler(w, http.StatusInternalServerError, err.Error())
			return
		}
		Log.Info("OK")

	case ErrIsNotMarkdown:
		Log.Info("Sending file (%s)...", wpath)
		v, err := wikiio.Load(wpath)
		if err != nil {
			Log.Error(err.Error())
			errorHandler(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(v.Raw())
		Log.Info("OK")

	default:
		Log.Error("%s (%s)", err, wpath)
		errorHandler(w, http.StatusNotFound, wpath)
		return
	}
}

func createPage(wpath string, w http.ResponseWriter, r *http.Request) {
	Log.Info("Creating new page (%s)...", wpath)

	switch err := wikiio.Create(wpath); err {
	case nil:
		// send success response
		errorHandler(w, http.StatusCreated, "")
		Log.Info("OK")
	case os.ErrExist:
		Log.Error(err.Error())
		errorHandler(w, http.StatusFound, "")
	default:
		Log.Error(err.Error())
		errorHandler(w, http.StatusInternalServerError, err.Error())
	}
}

func updatePage(wpath string, w http.ResponseWriter, r *http.Request) {
	backup, _ := strconv.ParseBool(r.URL.Query().Get("backup"))
	message, _ := url.QueryUnescape(r.URL.Query().Get("message"))
	name, _ := url.QueryUnescape(r.URL.Query().Get("name"))
	email, _ := url.QueryUnescape(r.URL.Query().Get("email"))

	if backup {
		Log.Info("Backing up (%s)...", wpath+".bac")
	} else {
		Log.Info("Saving (%s)...", wpath)
	}

	f, err := wikiio.Load(wpath)
	if err != nil {
		Log.Error(err.Error())
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
		Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Log.Info("Rendering preview...")
	// return generated html
	w.Write(parseMarkdown(b, filepath.Dir(wpath)))
	Log.Info("OK")
}

func renamePage(wpath string, w http.ResponseWriter, r *http.Request) {
	oldpath, err := url.QueryUnescape(r.URL.Query().Get("oldpath"))
	if err != nil {
		Log.Error("Unescape error: %s", err)
		code := http.StatusBadRequest
		http.Error(w, http.StatusText(code), code)
		return
	}

	Log.Info("Rename page (%s -> %s)...", oldpath, wpath)

	if err := wikiio.Rename(oldpath, wpath); err != nil {
		Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send success response
	w.Write(nil)
	Log.Info("OK")
}

func deletePage(wpath string, w http.ResponseWriter, r *http.Request) {
	Log.Info("Deleting page (%s)...", wpath)

	f, err := wikiio.Load(wpath)
	switch err {
	case nil:

	case os.ErrNotExist:
		Log.Error(err.Error())
		code := http.StatusNotFound
		http.Error(w, http.StatusText(code), code)
		return

	default:
		Log.Error("Load file error: %s", err)
		return
	}

	if err = f.Remove(); err != nil {
		Log.Error("Delete file error: %s", err)
		http.Error(w, "cannot delete file", http.StatusInternalServerError)
		return
	}

	// send success response
	w.Write(nil)
	Log.Info("Deleted")
}

func editorView(wpath string, w http.ResponseWriter, r *http.Request) {
	Log.Info("Editor requested (%s)", wpath)

	// check main dir
	f, _ := wikiio.Load(wpath)
	if f != nil {
		if f.DirMainPage() != nil {
			Log.Info("Requested file is dir, let's redirect to main file")
			http.Redirect(w, r,
				string(f.DirMainPage().URLPath())+"?view=editor", http.StatusFound)
			return
		} else if f.IsDir() {
			Log.Info("Requested file is dir, but main file dosen't exists.")
			errorHandler(w, http.StatusNotFound, wpath)
			return
		}
	}

	v, err := NewEditor(wpath)
	if err != nil {
		Log.Error(err.Error())
		errorHandler(w, http.StatusNotFound, wpath)
		return
	}

	// render html
	Log.Info("Rendering view...")
	err = v.Render(w)
	if err != nil {
		Log.Error("Rendering error!: %s", err)
		errorHandler(w, http.StatusInternalServerError, err.Error())
		return
	}
	Log.Info("OK")
}

func historyView(wpath string, w http.ResponseWriter, r *http.Request) {
	Log.Info("History view requested")

	v, err := LoadHistoryPage(wpath)
	if err != nil {
		Log.Error("%s (%s)", err, wpath)
		errorHandler(w, http.StatusNotFound, wpath)
		return
	}

	// render html
	Log.Info("Rendering history page (%s)...", wpath)
	err = v.Render(w)
	if err != nil {
		Log.Error("Rendering error!: %s", err)
		errorHandler(w, http.StatusInternalServerError, err.Error())
		return
	}
	Log.Info("OK")
}

func errorHandler(w http.ResponseWriter, status int, data string) {
	Log.Info("Rendering error view... (status: %d, data: %s)", status, data)

	w.WriteHeader(status)
	err := views.ExecuteTemplate(w, strconv.Itoa(status)+".html", data)
	if err != nil {
		// template not available
		http.Error(w, http.StatusText(status), status)
	}
	Log.Info("OK")
}
