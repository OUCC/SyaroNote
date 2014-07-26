package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

var (
	wikiRoot    string
	templateDir = "/home/yuntan/Workspace/go/src/syaro/templates/"
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

	filePath := req.URL.Path

	// FIXME wrong method (ext isn't always .md)
	if path.Ext(filePath) == "" {
		filePath += ".md"
	}

	if path.Ext(filePath) == ".md" {
		fmt.Println("Rendering page...")
		filePath = path.Clean(wikiRoot + filePath)

		// load md file
		page, err := LoadPage(filePath)
		if err != nil { // file not exist, so create new page
			fmt.Println("File not exist, create new page")
			page, err = NewPage(filePath)
			if err != nil { // strange error
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
		}

		// render html
		err = page.Render(rw)
		if err != nil {
			fmt.Println("rendering error!" + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

	} else if ext := path.Ext(filePath); ext == ".css" || ext == ".js" || ext == ".ico" {
		f, err := os.Open(path.Clean(templateDir + filePath))
		if err != nil {
			fmt.Println("not found")
			http.Error(rw, err.Error(), http.StatusNotFound)
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println("can't read")
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		fmt.Fprint(rw, string(b))
	}
}
