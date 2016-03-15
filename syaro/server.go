package main

import (
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/abbot/go-http-auth"

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
	muxRead := web.New()
	muxEdit := web.New()

	mux.Use(middleware.RequestID)
	if setting.verbose {
		mux.Use(middleware.Logger)
	}
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.AutomaticOptions)

	if setting.Auth.Mode != "" {
		// authenticator
		var a auth.AuthenticatorInterface
		if setting.Auth.Mode == "basic" {
			a = auth.NewBasicAuthenticator(setting.Auth.Realm, secret)
		} else if setting.Auth.Mode == "digest" {
			a = auth.NewDigestAuthenticator(setting.Auth.Realm, secret)
		}
		var authMiddleware = func(c *web.C, h http.Handler) http.Handler {
			return auth.JustCheck(a, h.ServeHTTP)
		}

		if !setting.Auth.PermitReadNobody {
			muxRead.Use(authMiddleware)
		}
		muxEdit.Use(authMiddleware)
	}

	// files under SYARO_DIR/public
	rootDir := http.Dir(filepath.Join(setting.syaroDir, PUBLIC_DIR))
	fileServer := http.StripPrefix(setting.urlPrefix, http.FileServer(rootDir))
	mux.Get(setting.urlPrefix+"/css/*", fileServer)
	mux.Get(setting.urlPrefix+"/fonts/*", fileServer)
	mux.Get(setting.urlPrefix+"/ico/*", fileServer)
	mux.Get(setting.urlPrefix+"/js/*", fileServer)
	mux.Get(setting.urlPrefix+"/images/*", fileServer)

	// readpnly APIs
	muxRead.Get(setting.urlPrefix+"/api/history", getHistory)
	muxRead.Get(setting.urlPrefix+"/api/get", getPage)

	// write APIs
	muxEdit.Get(setting.urlPrefix+"/api/new", createPage)
	muxEdit.Get(setting.urlPrefix+"/api/rename", renameFile)
	muxEdit.Get(setting.urlPrefix+"/api/delete", deleteFile)
	muxEdit.Post(setting.urlPrefix+"/api/update", updatePage)
	muxEdit.Post(setting.urlPrefix+"/api/upload", uploadFile)
	muxRead.Handle("/api/*", muxEdit)
	mux.Handle("/api/*", muxRead)

	// editor
	muxEdit.Get(setting.urlPrefix+"/edit",
		func(w http.ResponseWriter, r *http.Request) {
			tmpl.ExecuteTemplate(w, "editor.html", nil)
		})
	mux.Handle("/edit", muxEdit)

	muxRead.Get(setting.urlPrefix+"/search", searchHandler)
	muxRead.Get(setting.urlPrefix+"/*", mainHandler)
	mux.Handle("/*", muxRead)

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

func secret(user, realm string) string {
	if user == setting.Auth.User && realm == setting.Auth.Realm {
		return setting.Auth.Pass
	}
	return ""
}

/*
func authMiddleware(a auth.AuthenticatorInterface) interface{} {

	return func(c *web.C, h http.Handler) http.Handler {
		// return a.Wrap(func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		// 	h.ServeHTTP(w, &r.Request)
		// })
		return auth.JustCheck(a, h.ServeHTTP)
	}
}
*/
