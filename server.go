package main

import (
	"net"
	"net/http"
	"net/http/fcgi"
	"path/filepath"
	"strconv"
)

const (
	SYARO_PREFIX = "/syaro/"
)

func startServer() {
	// listen port
	l, err := net.Listen("tcp", ":"+strconv.Itoa(setting.port))
	if err != nil {
		loggerE.Fatalln("Error:", err)
	}

	mux := http.NewServeMux()

	// set http handler
	// for files under /syaro/
	mux.Handle(SYARO_PREFIX, http.StripPrefix(SYARO_PREFIX,
		http.FileServer(http.Dir(setting.tmplDir))))

	// for pages
	mux.HandleFunc(filepath.Clean("/"+setting.urlPrefix)+"/", handler)

	loggerM.Printf("Server started. Waiting connection localhost:%d/%s\n",
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
func handler(rw http.ResponseWriter, req *http.Request) {
	loggerM.Printf("Request received (%s)\n", req.URL.Path)

	path, err := filepath.Rel(filepath.Clean("/"+setting.urlPrefix),
		filepath.Clean("/"+req.URL.Path))
	if err != nil {
		loggerE.Println("Error:", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	path = filepath.Join(setting.wikiRoot, path)

	// load md file
	page, err := LoadPage(path)
	if err != nil {
		loggerE.Println("Error:", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// render html
	loggerM.Println("Rendering page...")
	err = page.Render(rw)
	if err != nil {
		loggerE.Println("Rendering error!", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
