package main

import (
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"net"
	"net/http"
	"net/http/fcgi"
	"path/filepath"
	"strconv"
)

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
	mux.Get(setting.urlPrefix+"/images/*", fileServer)

	mux.Get(setting.urlPrefix+"/api/get", getPage)
	mux.Get(setting.urlPrefix+"/api/new", createPage)
	mux.Get(setting.urlPrefix+"/api/rename", renameFile)
	mux.Get(setting.urlPrefix+"/api/delete", deleteFile)
	mux.Get(setting.urlPrefix+"/api/search", searchPage)
	mux.Post(setting.urlPrefix+"/api/update", updatePage)
	mux.Post(setting.urlPrefix+"/api/upload", uploadFile)

	mux.Get(setting.urlPrefix+"/*", mainHandler)
	mux.Get(setting.urlPrefix+"/edit",
		func(w http.ResponseWriter, r *http.Request) {
			tmpl.ExecuteTemplate(w, "editor.html", nil)
		})

	// listen port
	l, err := net.Listen("tcp", ":"+strconv.Itoa(setting.port))
	if err != nil {
		log.Fatal(err)
	}

	log.Notice("Server started. Waiting connection localhost:%d%s\n",
		setting.port, setting.urlPrefix)

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
