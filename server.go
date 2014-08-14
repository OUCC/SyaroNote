package main

import (
	"net"
	"net/http"
	"net/http/fcgi"
	"path/filepath"
	"strconv"
)

func startServer() {
	// listen port
	l, err := net.Listen("tcp", ":"+strconv.Itoa(setting.port))
	if err != nil {
		loggerE.Fatalln("Error:", err)
	}

	mux := http.NewServeMux()

	// fix url prefix
	if setting.urlPrefix != "" {
		setting.urlPrefix = filepath.Clean("/" + setting.urlPrefix)
	}

	// set http handler
	// files under SYARO_DIR/public
	rootDir := http.Dir(filepath.Join(setting.syaroDir, PUBLIC_DIR))
	fileServer := http.FileServer(rootDir)
	mux.Handle(setting.urlPrefix+"/css/", fileServer)
	mux.Handle(setting.urlPrefix+"/fonts/", fileServer)
	mux.Handle(setting.urlPrefix+"/ico/", fileServer)
	mux.Handle(setting.urlPrefix+"/img/", fileServer)
	mux.Handle(setting.urlPrefix+"/js/", fileServer)

	// for pages
	mux.HandleFunc(setting.urlPrefix+"/", handler)

	loggerM.Printf("Server started. Waiting connection localhost:%d%s\n",
		setting.port, setting.urlPrefix)
	loggerM.Println()

	if setting.fcgi {
		err = fcgi.Serve(l, mux)
	} else {
		err = http.Serve(l, mux)
	}

	if err != nil {
		loggerE.Fatal("Error:", err)
	}
}

// handler is basic http request handler
func handler(res http.ResponseWriter, req *http.Request) {
	loggerM.Printf("Request received (%s)\n", req.URL.Path)

	path, err := filepath.Rel(filepath.Clean("/"+setting.urlPrefix),
		filepath.Clean("/"+req.URL.Path))
	if err != nil {
		loggerE.Println("Error:", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	path = filepath.Join(setting.wikiRoot, path)

	// load md file
	page, err := LoadPage(path)
	if err != nil {
		loggerE.Println("Error:", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// render html
	loggerM.Println("Rendering page...")
	err = page.Render(res)
	if err != nil {
		loggerE.Println("Rendering error!", err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
