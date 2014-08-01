package main

import (
	"fmt"
	"os"
	"path"
)

const (
	TEMPLATE_DIR_DEFAULT_NAME = "templates"
	TEMPLATE_FILE_EXAMPLE     = "page.html"
	REPOSITORY_DIR            = "syaro"
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

	startServer()
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
		_, err := os.Stat(path.Join(os.Getenv("GOPATH"), "src", REPOSITORY_DIR,
			TEMPLATE_DIR_DEFAULT_NAME, TEMPLATE_FILE_EXAMPLE))
		if err == nil {
			return path.Join(os.Getenv("GOPATH"), "src", REPOSITORY_DIR,
				TEMPLATE_DIR_DEFAULT_NAME)
		}

		// second, /usr/local/share/syaro
		_, err = os.Stat(path.Join("/usr/local/share/syaro",
			TEMPLATE_DIR_DEFAULT_NAME, TEMPLATE_FILE_EXAMPLE))
		if err == nil {
			return path.Join("/usr/local/share/syaro", TEMPLATE_DIR_DEFAULT_NAME)
		}

		// can't find template dir
		return ""
	}
}
