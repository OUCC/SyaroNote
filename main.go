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
	Log.Notice("===== Syaro Wiki Server =====")
	Log.Notice("Starting...")
	Log.Notice("")

	findSyaroDir()
	if setting.SyaroDir == "" {
		Log.Fatal("Error: Can't find system file directory.")
	}

	Log.Notice("WikiName: %s", setting.WikiName)
	Log.Notice("WikiRoot: %s", setting.WikiRoot)
	Log.Notice("Syaro dir: %s", setting.SyaroDir)
	if setting.FCGI {
		Log.Notice("Fast CGI mode: ON")
	} else {
		Log.Notice("Fast CGI mode: OFF")
	}
	Log.Notice("Port: %d", setting.Port)
	Log.Notice("URL prefix: %s", setting.UrlPrefix)
	Log.Notice("MathJax: %t", setting.MathJax)
	Log.Notice("Highlight: %t", setting.Highlight)
	Log.Notice("Verbose output: %t", setting.Verbose)
	Log.Notice("")

	Log.Info("Parsing template...")
	err := setupViews()
	if err != nil {
		Log.Fatalf("Failed to parse template: %s", err)
	}
	Log.Info("Template parsed")

	Log.Info("Building file index...")
	wikiio.BuildIndex()
	Log.Debug("Index built")

	Log.Info("Setting up filesystem watcher...")
	wikiio.InitWatcher()
	defer wikiio.CloseWatcher()

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
			Log.Error("Can't find template file dir specified in argument")
			setting.SyaroDir = ""
			return
		}
	} else { // directory isn't specified by user so search it by myself
		paths := []string{
			".",
			"/usr/share/syaro",
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
		"mathjax":   func() bool { return setting.MathJax },
		"highlight": func() bool { return setting.Highlight },
	})
	tmpl, err := tmpl.ParseGlob(filepath.Join(setting.SyaroDir, VIEWS_DIR, "*.html"))
	if err != nil {
		return err
	}

	views = tmpl
	return nil
}
