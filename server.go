package main

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"

	"net"
	"net/http"
	"net/http/fcgi"
	"path/filepath"
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
	LoggerM.Printf("main.handler: Request received (%s)\n", req.URL.Path)

	wpath := filepath.Join(setting.WikiRoot,
		strings.TrimPrefix(req.URL.Path, setting.UrlPrefix))

	// load md file
	page, err := LoadPage(wpath)
	if err != nil {
		LoggerE.Println("Error: main.handler:", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// render html
	LoggerM.Println("main.handler: Rendering page...")
	err = page.Render(res)
	if err != nil {
		LoggerE.Println("Error: main.handler: Rendering error!", err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
