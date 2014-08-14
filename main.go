package main

import (
	"flag"
	"os"
	"path/filepath"
)

const (
	SYARO_REPOSITORY = "github.com/OUCC/syaro"
	PUBLIC_DIR       = "public"
	VIEWS_DIR        = "views"
)

var (
	setting *Setting
)

func main() {
	flag.Parse()
	setupLogger()

	// print welcome message
	loggerM.Println("===== Syaro Wiki Server =====")
	loggerM.Println("Starting...")
	loggerM.Println("")

	findSyaroDir()
	if setting.syaroDir == "" {
		loggerE.Fatalln("Error: Can't find system file directory.")
	}

	loggerM.Println("WikiRoot:", setting.wikiRoot)
	loggerM.Println("Syaro dir:", setting.syaroDir)
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
func findSyaroDir() {
	// if syaro dir is specified by user, search this dir
	if setting.syaroDir != "" {
		_, err := os.Stat(filepath.Join(setting.syaroDir, VIEWS_DIR))
		// if directory isn't exist
		if err != nil {
			loggerE.Println("Error: Can't find template file dir specified in argument")
			setting.syaroDir = ""
			return
		}
	} else { // directory isn't specified by user so search it by myself
		// first, $GOROOT/src/...
		path := filepath.Join(os.Getenv("GOPATH"), "src", SYARO_REPOSITORY)
		_, err := os.Stat(filepath.Join(path, VIEWS_DIR))
		if err == nil {
			setting.syaroDir = path
			return
		}

		// second, /usr/local/share/syaro
		path = "/usr/local/share/syaro"
		_, err = os.Stat(filepath.Join(path, VIEWS_DIR))
		if err == nil {
			setting.syaroDir = path
			return
		}

		// third, C:\Program Files\Syaro (Windows)
		path = "/Program Files/Syaro"
		_, err = os.Stat(filepath.Join(path, VIEWS_DIR))
		if err == nil {
			setting.syaroDir = path
			return
		}

		// can't find syaro dir
		setting.syaroDir = ""
		return
	}
}
