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
		LoggerE.Fatalln("Error: main.startServer:", err)
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
	mux.Handle(setting.UrlPrefix+"/img/", fileServer)
	mux.Handle(setting.UrlPrefix+"/js/", fileServer)
	mux.Handle(setting.UrlPrefix+"/lib/", fileServer)

	// for pages
	mux.HandleFunc(setting.UrlPrefix+"/", handler)

	LoggerM.Printf("main.startServer: Server started. Waiting connection localhost:%d%s\n",
		setting.Port, setting.UrlPrefix)
	LoggerM.Println()

	if setting.FCGI {
		err = fcgi.Serve(l, mux)
	} else {
		err = http.Serve(l, mux)
	}

	if err != nil {
		LoggerE.Fatal("Error: main.startServer:", err)
	}
}

// handler is basic http request handler
func handler(res http.ResponseWriter, req *http.Request) {
	requrl := req.URL

	LoggerM.Printf("main.handler: Request received (%s)\n", requrl.RequestURI())
	LoggerM.Printf("main.handler: Path: %s, Query: %s, Fragment: %s", requrl.Path,
		requrl.RawQuery, requrl.Fragment)

	wpath := strings.TrimPrefix(requrl.Path, setting.UrlPrefix)

	if re := regexp.MustCompile("^/error/\\d{3}$"); re.MatchString(wpath) {
		LoggerM.Println("main.handler: Error view requested")
		status, _ := strconv.Atoi(wpath[7:10])
		errorHandler(res, status, requrl.Query().Get("data"))
		return
	}

	switch requrl.Query().Get("view") {
	case "":
		switch requrl.Query().Get("action") {
		case "":
			LoggerM.Println("main.handler: Page requested")
			v, err := LoadPage(wpath)
			if err != nil {
				LoggerE.Println("Error: main.handler:", err)
				errorHandler(res, http.StatusNotFound, wpath)
				return
			}

			// render html
			LoggerM.Println("main.handler: Rendering view...")
			err = v.Render(res)
			if err != nil {
				LoggerE.Println("Error: main.handler: Rendering error!:", err)
				errorHandler(res, http.StatusInternalServerError, err.Error())
				return
			}

			LoggerV.Println("main.handler: Page rendered")

		case "create":
			LoggerM.Println("main.handler: Create new page")

			err := wikiio.Create(wpath)
			if err == os.ErrExist {
				LoggerE.Println("Error: main.handler: file already exists:", err)
				http.Error(res, "file already exists", http.StatusBadRequest)
				return
			}
			if err != nil {
				LoggerE.Println("Error: main.handler: Create file error:", err)
				http.Error(res, "cannot create file", http.StatusInternalServerError)
				return
			}

			// send success response
			res.Write(nil)
			LoggerV.Println("main.handler: New page created")

		case "rename":
			LoggerM.Println("main.handler: Rename page")

			oldpath, err := url.QueryUnescape(requrl.Query().Get("oldpath"))
			if err != nil {
				LoggerE.Println("Error: main.handler: Unescape error:", err)
				code := http.StatusBadRequest
				http.Error(res, http.StatusText(code), code)
				return
			}

			if err := wikiio.Rename(oldpath, wpath); err != nil {
				LoggerE.Println("Error: main.handler: Rename file error:", err)
				http.Error(res, "cannot rename file", http.StatusInternalServerError)
				return
			}

			// send success response
			res.Write(nil)
			LoggerV.Println("main.handler: Renamed")

		case "delete":
			LoggerM.Println("main.handler: Delete page")

			f, err := wikiio.Load(wpath)
			if err != nil {
				LoggerE.Println("Error: main.handler: Load file error:", err)
				code := http.StatusNotFound
				http.Error(res, http.StatusText(code), code)
				return
			}

			if err = f.Remove(); err != nil {
				LoggerE.Println("Error: main.handler: Delete file error:", err)
				http.Error(res, "cannot delete file", http.StatusInternalServerError)
				return
			}

			// send success response
			res.Write(nil)
			LoggerV.Println("main.handler: deleted")

		default:
			data := requrl.Query().Get("action")
			LoggerE.Printf("Error: main.handler: invalid URL query (action: %s)\n", data)
			status := http.StatusBadRequest
			http.Error(res, http.StatusText(status), status)
			return
		}

	case "editor":
		switch req.Method {
		case "", "GET":
			LoggerM.Println("main.handler: Editor requested")
			v, err := NewEditor(wpath)
			if err != nil {
				LoggerE.Println("Error: main.handler:", err)
				errorHandler(res, http.StatusNotFound, wpath)
				return
			}

			// render html
			LoggerM.Println("main.handler: Rendering view...")
			err = v.Render(res)
			if err != nil {
				LoggerE.Println("Error: main.handler: Rendering error!:", err)
				errorHandler(res, http.StatusInternalServerError, err.Error())
				return
			}

		case "POST":
			LoggerM.Println("main.handler: Save requested")
			f, err := wikiio.Load(wpath)
			if err != nil {
				LoggerE.Println("Error: main.handler:", err)
				http.Error(res, wpath+"not found", http.StatusNotFound)
				return
			}

			b, _ := ioutil.ReadAll(req.Body)
			err = f.Save(b)
			if err != nil {
				LoggerE.Println("Error: main.handler: couldn't write:", err)
				http.Error(res, "cannot save document", http.StatusInternalServerError)
				return
			}
			LoggerM.Println("main.handler: File saved")

			// send success response
			res.Write(nil)
		}

	case "history":
		LoggerM.Println("main.handler: History view requested")

		// not implemented
		errorHandler(res, http.StatusNotImplemented, "History")
		return

	default:
		data := requrl.Query().Get("view")
		LoggerE.Printf("Error: main.handler: invalid URL query (view: %s)\n", data)
		errorHandler(res, http.StatusNotFound, data)
		return
	}
}

func errorHandler(res http.ResponseWriter, status int, data string) {
	LoggerV.Println("main.errorHandler: Rendering error view... ",
		"(status:", status, ", data:", data, ")")

	err := views.ExecuteTemplate(res, strconv.Itoa(status)+".html", data)
	if err != nil {
		// template not available
		LoggerE.Println("Error: main.errorHandler:", err)
		http.Error(res, http.StatusText(status), status)
	}
}
