package main

import (
	"fmt"
	"net/http"
	"os"
)

var (
	wikiRoot string
)

func main() {
	// print welcome message
	fmt.Println("=== syaro wiki server ===")
	fmt.Println("Starting...")

	// set repository root
	if len(os.Args) == 2 {
		wikiRoot = os.Args[1]
	} else {
		wikiRoot = "./"
	}
	fmt.Println("WikiRoot:" + wikiRoot)

	fmt.Println("Server started. Waiting connection on port :8080")
	fmt.Println()

	// set http handler
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Response received (%s)\n", req.URL.Path)

	// generate html
	filePath := wikiRoot + req.URL.Path
	page, err := LoadPage(filePath)
	if err != nil { // file not exist, so create new page
		page, err = NewPage(filePath)
		if err != nil { // strange error
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}

	// send html
	fmt.Fprint(rw, string(page.HTMLBody))
}
