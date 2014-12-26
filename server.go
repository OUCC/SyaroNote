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
		Log.Fatalf("main.startServer: %s", err)
	}
}

// handler is basic http request handler
func handler(res http.ResponseWriter, req *http.Request) {
	requrl := req.URL

	Log.Info("Request received (%s)", requrl.RequestURI())

	// url unescape (+ -> <Space>)
	requrl.Path = strings.Replace(requrl.Path, "+", " ", -1)

	Log.Info("Path: %s, Query: %s, Fragment: %s", requrl.Path,
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
			Log.Info("Page requested")
			v, err := LoadPage(wpath)
			switch err {
			case nil:
				// render html
				Log.Info("Rendering view...")
				err = v.Render(res)
				if err != nil {
					Log.Error("Rendering error!: %s", err)
					errorHandler(res, http.StatusInternalServerError, err.Error())
					return
				}
				Log.Info("Page rendered")

			case ErrIsNotMarkdown:
				Log.Info("File requested")
				v, err := wikiio.Load(wpath)
				if err != nil {
					Log.Error(err.Error())
					http.Error(res, err.Error(), http.StatusNotFound)
					return
				}
				res.Write(v.Raw())
				Log.Info("File sent")

			default:
				Log.Error(err.Error())
				errorHandler(res, http.StatusNotFound, wpath)
				return
			}

		case "create":
			Log.Info("Create new page")

			err := wikiio.Create(wpath)
			if err == os.ErrExist {
				Log.Error("file already exists (%s)", err)
				http.Error(res, "file already exists", http.StatusBadRequest)
				return
			}
			if err != nil {
				Log.Error("Create file error: %s", err)
				http.Error(res, "cannot create file", http.StatusInternalServerError)
				return
			}

			// send success response
			res.Write(nil)
			Log.Info("New page created")

		case "rename":
			Log.Info("Rename page")

			oldpath, err := url.QueryUnescape(requrl.Query().Get("oldpath"))
			if err != nil {
				Log.Error("Unescape error: %s", err)
				code := http.StatusBadRequest
				http.Error(res, http.StatusText(code), code)
				return
			}

			if err := wikiio.Rename(oldpath, wpath); err != nil {
				Log.Error("Rename file error: %s", err)
				http.Error(res, "cannot rename file", http.StatusInternalServerError)
				return
			}

			// send success response
			res.Write(nil)
			Log.Info("Renamed")

		case "delete":
			Log.Info("Delete page")

			f, err := wikiio.Load(wpath)
			if err != nil {
				Log.Error("Load file error: %s", err)
				code := http.StatusNotFound
				http.Error(res, http.StatusText(code), code)
				return
			}

			if f.IsDir() {
				err = f.RemoveAll()
			} else {
				err = f.Remove()
			}
			if err != nil {
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
			Log.Info("Editor requested")

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

		case "POST":
			Log.Info("Save requested")
			f, err := wikiio.Load(wpath)
			if err != nil {
				Log.Error(err.Error())
				http.Error(res, wpath+"not found", http.StatusNotFound)
				return
			}

			b, _ := ioutil.ReadAll(req.Body)
			err = f.Save(b)
			if err != nil {
				Log.Error("couldn't write: %s", err)
				http.Error(res, "cannot save document", http.StatusInternalServerError)
				return
			}
			Log.Info("File saved")

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
	Log.Info("Request received ReuqestURI: %s", req.RequestURI)

	path := req.URL.Query().Get("path")
	dir := filepath.Dir(path)
	Log.Info("dir: %s", dir)

	// raw markdown
	text, _ := ioutil.ReadAll(req.Body)

	// return generated html
	res.Write(parseMarkdown(text, dir))
	Log.Info("Response sent")
}

func errorHandler(res http.ResponseWriter, status int, data string) {
	Log.Info("Rendering error view... (status: %d, data: %s)", status, data)

	err := views.ExecuteTemplate(res, strconv.Itoa(status)+".html", data)
	if err != nil {
		// template not available
		Log.Error(err.Error())
		http.Error(res, http.StatusText(status), status)
	}
}
