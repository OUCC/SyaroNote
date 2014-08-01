package main

import (
	"fmt"
	"net/http"
	"path/filepath"
)

const (
	SYARO_PREFIX = "/syaro/"
)

func startServer() {
	fmt.Println("Server started. Waiting connection on port :8080")
	fmt.Println()

	// set http handler
	// for files under /syaro/
	http.Handle(SYARO_PREFIX, http.StripPrefix(SYARO_PREFIX,
		http.FileServer(http.Dir(templateDir))))

	// for pages
	http.HandleFunc("/", handler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// handler is basic http request handler
func handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Request received (%s)\n", req.URL.Path)

	path := filepath.Join(wikiRoot, req.URL.Path)

	// load md file
	page, err := LoadPage(path)
	if err != nil {
		fmt.Println("Error:", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// render html
	fmt.Println("Rendering page...")
	err = page.Render(rw)
	if err != nil {
		fmt.Println("Rendering error!", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
