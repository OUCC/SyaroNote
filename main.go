package main

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/wikiio"

	"flag"
	"html/template"
	"os"
	"path/filepath"
)

const (
	SYARO_REPOSITORY = "github.com/OUCC/syaro"
	PUBLIC_DIR       = "public"
	VIEWS_DIR        = "views"
)

var (
	views *template.Template
)

func main() {
	flag.Parse()
	SetupLogger()

	// print welcome message
	LoggerM.Println("===== Syaro Wiki Server =====")
	LoggerM.Println("Starting...")
	LoggerM.Println("")

	findSyaroDir()
	if setting.SyaroDir == "" {
		LoggerE.Fatalln("Error: Can't find system file directory.")
	}

	LoggerM.Println("WikiName:", setting.WikiName)
	LoggerM.Println("WikiRoot:", setting.WikiRoot)
	LoggerM.Println("Syaro dir:", setting.SyaroDir)
	if setting.FCGI {
		LoggerM.Println("Fast CGI mode: YES")
	} else {
		LoggerM.Println("Fast CGI mode: NO")
	}
	LoggerM.Println("Port:", setting.Port)
	LoggerM.Println("URL prefix:", setting.UrlPrefix)
	LoggerM.Println("")

	LoggerM.Println("Parsing template...")
	err := setupViews()
	if err != nil {
		LoggerE.Fatalln("Failed to parse template:", err)
	}
	LoggerM.Println("Template parsed")

	LoggerM.Println("Building index...")
	wikiio.BuildIndex()
	LoggerM.Println("Index built")

	startServer()
}

// findTemplateDir finds template directory contains html, css, etc...
// dir is directory specified by user as template dir.
// This search several directory and return right dir.
// If not found, return empty string.
func findSyaroDir() {
	// if syaro dir is specified by user, search this dir
	if setting.SyaroDir != "" {
		_, err := os.Stat(filepath.Join(setting.SyaroDir, VIEWS_DIR))
		// if directory isn't exist
		if err != nil {
			LoggerE.Println("Error: Can't find template file dir specified in argument")
			setting.SyaroDir = ""
			return
		}
	} else { // directory isn't specified by user so search it by myself
		paths := []string{
			".",
			"/usr/local/share/syaro",
			"/Program Files/Syaro",
		}

		for _, path := range paths {
			_, err := os.Stat(filepath.Join(path, VIEWS_DIR))
			if err == nil {
				setting.SyaroDir = path
				return
			}
		}

		// can't find syaro dir
		setting.SyaroDir = ""
		return
	}
}

func setupViews() error {
	// funcs for template
	tmpl := template.New("").Funcs(template.FuncMap{
		"add":       func(a, b int) int { return a + b },
		"wikiName":  func() string { return setting.WikiName },
		"urlPrefix": func() string { return setting.UrlPrefix },
	})
	tmpl, err := tmpl.ParseGlob(filepath.Join(setting.SyaroDir, VIEWS_DIR, "*.html"))
	if err != nil {
		return err
	}

	views = tmpl
	return nil
}
