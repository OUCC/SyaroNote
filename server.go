package main

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"

	"net"
	"net/http"
	"net/http/fcgi"
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

	LoggerM.Printf("main.handler: Request received (%s)\n", requrl.Path)
	LoggerM.Printf("main.handler: Path: %s, Query: %s, Fragment: %s", requrl.Path,
		requrl.RawQuery, requrl.Fragment)

	// TODO url unescape
	wpath := strings.TrimPrefix(requrl.Path, setting.UrlPrefix)

	if re := regexp.MustCompile("^/error/\\d{3}$"); re.MatchString(wpath) {
		LoggerM.Println("main.handler: Error view requested")
		status, _ := strconv.Atoi(wpath[7:10])
		errorHandler(res, status, requrl.Query().Get("data"))
		return
	}

	var v View
	var err error

	switch requrl.Query().Get("view") {
	case "":
		switch requrl.Query().Get("action") {
		case "":
			LoggerM.Println("main.handler: Page requested")
			v, err = LoadPage(wpath)

		case "new":
			LoggerM.Println("main.handler: Create new page")

			// not implemented
			errorHandler(res, http.StatusNotImplemented, "Editor")
			return

		case "rename":
			LoggerM.Println("main.handler: Rename page")

			// not implemented
			errorHandler(res, http.StatusNotImplemented, "Editor")
			return

		case "delete":
			LoggerM.Println("main.handler: Delete page")

			// not implemented
			errorHandler(res, http.StatusNotImplemented, "Editor")
			return

		default:
			data := requrl.Query().Get("action")
			LoggerE.Printf("Error: main.handler: invalid URL query (action: %s)\n", data)
			errorHandler(res, http.StatusNotFound, data)
			return
		}

	case "editor":
		LoggerM.Println("main.handler: Editor requested")
		v, err = NewEditor(wpath)

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

	if err != nil {
		LoggerE.Println("Error: main.handler:", err)
		errorHandler(res, http.StatusNotFound, wpath)
		return
	}

	// render html
	LoggerM.Println("main.handler: Rendering view...")
	err = v.Render(res)
	if err != nil {
		LoggerE.Println("Error: main.handler: Rendering error!", err)
		errorHandler(res, http.StatusInternalServerError, err.Error())
		return
	}
}

func errorHandler(res http.ResponseWriter, status int, data string) {
	LoggerV.Println("main.errorHandler: Rendering error view... ",
		"(status:", status, ", data:", data, ")")

	err := views.ExecuteTemplate(res, strconv.Itoa(status)+".html", data)
	if err != nil {
		LoggerE.Println("Error: main.errorHandler:", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
