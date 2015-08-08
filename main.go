package main

import (
	"flag"
	"html/template"
	"os"
	"path/filepath"
	"time"
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
	log.Notice("===== Syaro Wiki Server =====")
	log.Notice("Starting...")
	log.Notice("")

	findSyaroDir()
	if setting.SyaroDir == "" {
		log.Fatal("Error: Can't find system file directory.")
	}

	log.Notice("WikiName: %s", setting.wikiName)
	log.Notice("WikiRoot: %s", setting.wikiRoot)
	log.Notice("Syaro dir: %s", setting.SyaroDir)
	if setting.fcgi {
		log.Notice("Fast CGI mode: ON")
	} else {
		log.Notice("Fast CGI mode: OFF")
	}
	log.Notice("Port: %d", setting.port)
	log.Notice("URL prefix: %s", setting.urlPrefix)
	setting.gitMode = CheckRepository()
	if setting.gitMode {
		log.Notice("Git mode: ON")
	} else {
		log.Notice("Git mode: OFF")
	}
	//log.Notice("MathJax: %t", setting.mathjax)
	//log.Notice("Highlight: %t", setting.highlight)
	log.Notice("Verbose output: %t", setting.verbose)
	log.Notice("")

	log.Info("Parsing template...")
	err := setupViews()
	if err != nil {
		log.Fatalf("Failed to parse template: %s", err)
	}
	log.Info("Template parsed")

	log.Info("Setting up filesystem watcher...")
	InitWatcher()
	defer CloseWatcher()

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
			log.Error("Can't find template file dir specified in argument")
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
		"op":        OpString,
		"timef":     func(t time.Time) string { return t.Format("Mon _2 Jan 2006") },
		"wikiName":  func() string { return setting.wikiName },
		"urlPrefix": func() string { return setting.urlPrefix },
		"mathjax":   func() bool { return setting.mathjax },
		"highlight": func() bool { return setting.highlight },
		"gitmode":   func() bool { return setting.gitMode },
		"byteToStr": func(b []byte) string { return string(b) },
	})
	tmpl, err := tmpl.ParseGlob(filepath.Join(setting.SyaroDir, VIEWS_DIR, "*.html"))
	if err != nil {
		return err
	}

	views = tmpl
	return nil
}
