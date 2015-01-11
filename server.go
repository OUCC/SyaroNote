package main

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/wikiio"

	"io/ioutil"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func startServer() {
	// listen port
	l, err := net.Listen("tcp", ":"+strconv.Itoa(setting.Port))
	if err != nil {
		Log.Fatal(err)
	}

	mux := http.NewServeMux()

	// fix url prefix
	if setting.UrlPrefix != "" {
		setting.UrlPrefix = filepath.Clean("/" + setting.UrlPrefix)
	}

	// set http handler
	// files under SYARO_DIR/public
	rootDir := http.Dir(filepath.Join(setting.SyaroDir, PUBLIC_DIR))
	fileServer := http.StripPrefix(setting.UrlPrefix, http.FileServer(rootDir))
	mux.Handle(setting.UrlPrefix+"/css/", fileServer)
	mux.Handle(setting.UrlPrefix+"/fonts/", fileServer)
	mux.Handle(setting.UrlPrefix+"/ico/", fileServer)
	mux.Handle(setting.UrlPrefix+"/js/", fileServer)
	mux.Handle(setting.UrlPrefix+"/lib/", fileServer)

	// for editor preview
	mux.HandleFunc(setting.UrlPrefix+"/preview", previewHandler)
	// for pages
	mux.HandleFunc(setting.UrlPrefix+"/", handler)

	Log.Notice("Server started. Waiting connection localhost:%d%s\n",
		setting.Port, setting.UrlPrefix)

	if setting.FCGI {
		err = fcgi.Serve(l, mux)
	} else {
		err = http.Serve(l, mux)
	}

	if err != nil {
		Log.Fatal(err)
	}
}

// handler is basic http request handler
func handler(res http.ResponseWriter, req *http.Request) {
	requrl := req.URL

	Log.Debug("Request received (%s)", requrl.RequestURI())

	// url unescape (+ -> <Space>)
	requrl.Path = strings.Replace(requrl.Path, "+", " ", -1)

	Log.Debug("Path: %s, Query: %s, Fragment: %s", requrl.Path,
		requrl.RawQuery, requrl.Fragment)

	wpath := strings.TrimPrefix(requrl.Path, setting.UrlPrefix)

	if re := regexp.MustCompile("^/error/\\d{3}$"); re.MatchString(wpath) {
		Log.Info("Error view requested")
		status, _ := strconv.Atoi(wpath[7:10])
		errorHandler(res, status, requrl.Query().Get("data"))
		return
	}

	switch requrl.Query().Get("view") {
	case "":
		switch requrl.Query().Get("action") {
		case "":
			v, err := LoadPage(wpath)
			switch err {
			case nil:
				// render html
				Log.Info("Rendering page (%s)...", wpath)
				err = v.Render(res)
				if err != nil {
					Log.Error("Rendering error!: %s", err)
					errorHandler(res, http.StatusInternalServerError, err.Error())
					return
				}
				Log.Info("OK")

			case ErrIsNotMarkdown:
				Log.Info("Sending file (%s)...", wpath)
				v, err := wikiio.Load(wpath)
				if err != nil {
					Log.Error(err.Error())
					http.Error(res, err.Error(), http.StatusNotFound)
					return
				}
				res.Write(v.Raw())
				Log.Info("OK")

			default:
				Log.Error("%s (%s)", err, wpath)
				errorHandler(res, http.StatusNotFound, wpath)
				return
			}

		case "create":
			Log.Info("Creating new page (%s)...", wpath)

			switch err := wikiio.Create(wpath); err {
			case nil:
				// send success response
				res.Write(nil)
				Log.Info("OK")

			case os.ErrExist:
				Log.Error(err.Error())
				http.Error(res, err.Error(), http.StatusBadRequest)

			default:
				Log.Error(err.Error())
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}

		case "rename":
			oldpath, err := url.QueryUnescape(requrl.Query().Get("oldpath"))
			if err != nil {
				Log.Error("Unescape error: %s", err)
				code := http.StatusBadRequest
				http.Error(res, http.StatusText(code), code)
				return
			}

			Log.Info("Rename page (%s -> %s)...", oldpath, wpath)

			if err := wikiio.Rename(oldpath, wpath); err != nil {
				Log.Error(err.Error())
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}

			// send success response
			res.Write(nil)
			Log.Info("OK")

		case "delete":
			Log.Info("Deleting page (%s)...", wpath)

			f, err := wikiio.Load(wpath)
			switch err {
			case nil:

			case os.ErrNotExist:
				Log.Error(err.Error())
				code := http.StatusNotFound
				http.Error(res, http.StatusText(code), code)
				return

			default:
				Log.Error("Load file error: %s", err)
				return
			}

			if err = f.Remove(); err != nil {
				Log.Error("Delete file error: %s", err)
				http.Error(res, "cannot delete file", http.StatusInternalServerError)
				return
			}

			// send success response
			res.Write(nil)
			Log.Info("Deleted")

		default:
			data := requrl.Query().Get("action")
			Log.Error("Error: main.handler: invalid URL query (action: %s)", data)
			status := http.StatusBadRequest
			http.Error(res, http.StatusText(status), status)
			return
		}

	case "editor":
		switch req.Method {
		case "", "GET":
			Log.Info("Editor requested (%s)", wpath)

			// check main dir
			f, _ := wikiio.Load(wpath)
			if f != nil {
				if f.DirMainPage() != nil {
					Log.Info("Requested file is dir, let's redirect to main file")
					http.Redirect(res, req,
						string(f.DirMainPage().URLPath())+"?view=editor", http.StatusFound)
					return
				} else if f.IsDir() {
					Log.Info("Requested file is dir, but main file dosen't exists.")
					Log.Info("Return 404 page")
					errorHandler(res, http.StatusNotFound, wpath)
					return
				}
			}

			v, err := NewEditor(wpath)
			if err != nil {
				Log.Error(err.Error())
				errorHandler(res, http.StatusNotFound, wpath)
				return
			}

			// render html
			Log.Info("Rendering view...")
			err = v.Render(res)
			if err != nil {
				Log.Error("Rendering error!: %s", err)
				errorHandler(res, http.StatusInternalServerError, err.Error())
				return
			}
			Log.Info("OK")

		case "POST":
			Log.Info("Saving (%s)...", wpath)
			message, _ := url.QueryUnescape(requrl.Query().Get("message"))
			name, _ := url.QueryUnescape(requrl.Query().Get("name"))
			email, _ := url.QueryUnescape(requrl.Query().Get("email"))

			Log.Info("message: %s", message)
			Log.Info("author name: %s, email: %s", name, email)

			f, err := wikiio.Load(wpath)
			if err != nil {
				Log.Error(err.Error())
				http.Error(res, wpath+"not found", http.StatusNotFound)
				return
			}

			b, _ := ioutil.ReadAll(req.Body)
			defer req.Body.Close()
			err = f.Save(b, message, name, email)
			if err != nil {
				Log.Error(err.Error())
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			Log.Info("OK")

			// send success response
			res.Write(nil)
		}

	case "history":
		Log.Info("History view requested")

		// not implemented
		errorHandler(res, http.StatusNotImplemented, "History")
		return

	default:
		data := requrl.Query().Get("view")
		Log.Error("Error: main.handler: invalid URL query (view: %s)", data)
		errorHandler(res, http.StatusNotFound, data)
		return
	}
}

// previewHandler for markdown preview in editor
func previewHandler(res http.ResponseWriter, req *http.Request) {
	Log.Info("Rendering preview...")

	path := req.URL.Query().Get("path")
	dir := filepath.Dir(path)
	Log.Debug("dir: %s", dir)

	// raw markdown
	text, _ := ioutil.ReadAll(req.Body)

	// return generated html
	res.Write(parseMarkdown(text, dir))
	Log.Info("OK")
}

func errorHandler(res http.ResponseWriter, status int, data string) {
	Log.Info("Rendering error view... (status: %d, data: %s)", status, data)

	err := views.ExecuteTemplate(res, strconv.Itoa(status)+".html", data)
	if err != nil {
		// template not available
		Log.Error(err.Error())
		http.Error(res, http.StatusText(status), status)
	}
	Log.Info("OK")
}
