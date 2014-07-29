package syaro

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

const (
	TEMPLATE_FILE_EXAMPLE = "page.html"
	REPOSITORY_DIR        = "syaro"
)

var (
	wikiRoot    string
	templateDir string
)

func main() {
	// print welcome message
	fmt.Println("===== Syaro Wiki Server =====")
	fmt.Println("Starting...")

	templateDir = findTemplateDir("")
	if templateDir == "" {
		fmt.Println("Error: Can't find template dir.")
		return
	}
	fmt.Println("Template dir:", templateDir)

	// set repository root
	if len(os.Args) == 2 {
		wikiRoot = os.Args[1]
	} else {
		wikiRoot = "./"
	}
	fmt.Println("WikiRoot:", wikiRoot)

	fmt.Println("Server started. Waiting connection on port :8080")
	fmt.Println()

	// set http handler
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

// findTemplateDir finds template directory contains html, css, etc...
// dir is directory specified by user as template dir.
// This search several directory and return right dir.
// If not found, return empty string.
func findTemplateDir(dir string) string {
	// if template dir is specified by user, search this dir
	if dir != "" {
		_, err := os.Stat(path.Join(dir, TEMPLATE_FILE_EXAMPLE))
		// if directory isn't exist
		if err != nil {
			fmt.Println("Error: Can't find template file dir specified in argument")
			return ""
		}
		return dir
	} else { // directory isn't specified by user so search it by myself
		// first, $GOROOT/src/...
		_, err := os.Stat(path.Join(os.Getenv("GOROOT"), "src", REPOSITORY_DIR,
			"template", TEMPLATE_FILE_EXAMPLE))
		if err == nil {
			return path.Join(os.Getenv("GOROOT"), "src", REPOSITORY_DIR)
		}

		// second, /usr/local/share/syaro
		_, err = os.Stat(path.Join("/usr/local/share/syaro", "templates",
			TEMPLATE_FILE_EXAMPLE))
		if err == nil {
			return path.Join("/usr/local/share/syaro", "templates")
		}

		// can't find template dir
		return ""
	}
}

// handler is basic http request handler
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
			fmt.Println("rendering error!", err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

	} else if ext := path.Ext(filePath); ext == ".css" || ext == ".js" {
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
