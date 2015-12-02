package main

import (
	pb "github.com/OUCC/syaro/gitservice"
	"github.com/OUCC/syaro/markdown"

	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const (
	PUBLIC_DIR   = "public"
	TEMPLATE_DIR = "template"
	ENV_SYARODIR = "SYARODIR"
)

var (
	version string
	tmpl    *template.Template
)

func main() {
	parseFlags()
	setupLogger()

	// print welcome message
	log.Notice("===== Syaro Wiki Server %s =====", version)
	log.Notice("Starting...")
	log.Notice("")

	findsyaroDir()
	if setting.syaroDir == "" {
		log.Fatal("Error: Can't find template file directory.")
	}

	log.Notice("WikiRoot: %s", setting.wikiRoot)
	log.Notice("Syaro dir: %s", setting.syaroDir)
	if setting.fcgi {
		log.Notice("Fast CGI mode: ON")
	} else {
		log.Notice("Fast CGI mode: OFF")
	}
	log.Notice("Port: %d", setting.port)
	log.Notice("URL prefix: %s", setting.urlPrefix)
	if setting.gitMode {
		log.Info("Loading Git plugin...")
		cmd := exec.Command(filepath.Join(setting.syaroDir, "gitplugin"),
			":"+strconv.Itoa(setting.port+1), setting.wikiRoot)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		go func() {
			if err := cmd.Run(); err != nil { // blocking
				log.Fatalf("Git plugin unexpectedly crashed: %s", err)
			}
			log.Fatal("Git plugin unexpectedly crashed")
		}()
		defer cmd.Process.Kill()

		log.Notice("Git mode: ON")
	} else {
		log.Notice("Git mode: OFF")
	}
	log.Notice("MathJax: %t", setting.mathjax)
	log.Notice("Highlight: %t", setting.highlight)
	log.Notice("Verbose output: %t\n", setting.verbose)

	log.Info("Parsing template...")
	err := setupViews()
	if err != nil {
		log.Fatalf("Failed to parse template: %s", err)
	}
	log.Info("Template parsed")

	log.Info("Setting up fs watcher...")
	go fsWatcher()

	log.Info("Setting up file index...")
	go idxBuilder()
	refresh <- "/"

	log.Info("Setting up websockets...")
	// TODO

	// set link worker
	markdown.LinkWorker = linkWorker

	startServer()
}

// findTemplateDir finds template directory contains html, css, etc...
// dir is directory specified by user as template dir.
// This search several directory and return right dir.
// If not found, return empty string.
func findsyaroDir() {
	// if syaro dir is specified by user, search this dir
	env := os.Getenv(ENV_SYARODIR)
	if env != "" {
		_, err := os.Stat(filepath.Join(env, TEMPLATE_DIR))
		// if directory isn't exist
		if err != nil {
			log.Error("Can't find template file dir specified by env")
			setting.syaroDir = ""
			return
		}
		setting.syaroDir = env
	} else { // directory isn't specified by user so search it by myself
		paths := []string{
			".",
			"/usr/share/syaro",
			"/usr/local/share/syaro",
			`\Program Files\Syaro`,
		}

		for _, path := range paths {
			_, err := os.Stat(filepath.Join(path, TEMPLATE_DIR))
			if err == nil {
				setting.syaroDir = path
				return
			}
		}

		// can't find syaro dir
		setting.syaroDir = ""
		return
	}
}

func setupViews() error {
	// funcs for template
	tmpl = template.New("").Funcs(template.FuncMap{
		"op": func(op pb.Change_Op) string {
			switch op {
			case pb.Change_OpAdd:
				return "Add"
			case pb.Change_OpUpdate:
				return "Edit"
			case pb.Change_OpRename:
				return "Rename"
			}
			return ""
		},
		"timef":     func(t time.Time) string { return t.Format("Mon _2 Jan 2006") },
		"urlPrefix": func() string { return setting.urlPrefix },
		"mathjax":   func() bool { return setting.mathjax },
		"highlight": func() bool { return setting.highlight },
		"emoji":     func() bool { return setting.emoji },
		"gitmode":   func() bool { return setting.gitMode },
		"search":    func() bool { return setting.search },
	})
	var err error
	tmpl, err = tmpl.ParseGlob(filepath.Join(setting.syaroDir, TEMPLATE_DIR, "*.html"))
	if err != nil {
		return err
	}
	return nil
}
