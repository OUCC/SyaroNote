package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	SYARO_REPOSITORY = "github.com/OUCC/syaro"
)

var (
	setting *Setting
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--usage", "--help", "-h":
			flag.Usage()
		}
	}

	// print welcome message
	fmt.Println("===== Syaro Wiki Server =====")
	fmt.Println("Starting...")
	fmt.Println("")

	flag.Parse()

	setupLogger()

	findTemplateDir()
	if setting.tmplDir == "" {
		loggerM.Println("Error: Can't find template dir.")
		loggerM.Println("Server shutdown.")
		return
	}

	loggerM.Println("WikiRoot:", setting.wikiRoot)
	loggerM.Println("Template dir:", setting.tmplDir)
	if setting.fcgi {
		loggerM.Println("Fast CGI mode: YES")
	} else {
		loggerM.Println("Fast CGI mode: NO")
	}
	loggerM.Println("Port:", setting.port)
	loggerM.Println("URL prefix:", setting.urlPrefix)
	loggerM.Println("")

	startServer()
}

// findTemplateDir finds template directory contains html, css, etc...
// dir is directory specified by user as template dir.
// This search several directory and return right dir.
// If not found, return empty string.
func findTemplateDir() {
	const (
		TEMPLATE_DIR_DEFAULT_NAME = "templates"
		TEMPLATE_FILE_EXAMPLE     = "page.html"
	)

	// if template dir is specified by user, search this dir
	if setting.tmplDir != "" {
		_, err := os.Stat(filepath.Join(setting.tmplDir, TEMPLATE_FILE_EXAMPLE))
		// if directory isn't exist
		if err != nil {
			loggerE.Println("Error: Can't find template file dir specified in argument")
			setting.tmplDir = ""
			return
		}
	} else { // directory isn't specified by user so search it by myself
		// first, $GOROOT/src/...
		path := filepath.Join(os.Getenv("GOPATH"), "src", SYARO_REPOSITORY,
			TEMPLATE_DIR_DEFAULT_NAME)
		_, err := os.Stat(filepath.Join(path, TEMPLATE_FILE_EXAMPLE))
		if err == nil {
			setting.tmplDir = path
			return
		}

		// second, /usr/local/share/syaro
		path = filepath.Join("/usr/local/share/syaro",
			TEMPLATE_DIR_DEFAULT_NAME)
		_, err = os.Stat(filepath.Join(path, TEMPLATE_FILE_EXAMPLE))
		if err == nil {
			setting.tmplDir = path
			return
		}

		// can't find template dir
		setting.tmplDir = ""
		return
	}
}
