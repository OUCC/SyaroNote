package main

import (
	"gopkg.in/yaml.v2"

	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const yamlPath = ".syaronote.yaml"

var setting struct {
	// server
	syaroDir   string
	wikiRoot   string
	port       int
	urlPrefix  string
	fcgi       bool
	singleFile bool // TODO
	readonly   bool // TODO
	gitMode    bool
	search     bool

	// console output
	verbose bool
	quiet   bool
	color   bool

	// authentication
	Auth struct { // TODO
		Reader string `yaml:"reader"`
		Writer string `yaml:"writer"`
	} `yaml:"auth"`

	// markdown related
	Markdown struct {
		MathJax   bool `yaml:"mathjax"`
		Highlight bool `yaml:"highlight"`
		Emoji     bool `yaml:"emoji"`
	} `yaml:"markdown"`

	// post action
	Actions struct {
		// post actions
		PostCreate []string `yaml:"post_create"`
		PostUpdate []string `yaml:"post_update"`
		PostRename []string `yaml:"post_rename"`
		PostDelete []string `yaml:"post_remove"`
		PostUpload []string `yaml:"post_upload"`
	} `yaml:"actions"`
}

func init() {
	// setting up flags
	flag.IntVar(&setting.port, "port", 8080,
		"Port number")
	// flag.StringVar(&setting.urlPrefix, "url-prefix", "",
	// 	"URL prefix (ex. if prefix is syarowiki, URL is localhost:PORT/syarowiki/)")
	flag.BoolVar(&setting.fcgi, "fcgi", false,
		"If true, syaro runs on fast cgi mode")
	flag.BoolVar(&setting.singleFile, "single", false,
		"Single file mode")
	flag.BoolVar(&setting.gitMode, "gitmode", false,
		"Enable git integration")
	flag.BoolVar(&setting.readonly, "readonly", false,
		"Readonly mode")
	flag.BoolVar(&setting.search, "search", false,
		"enable indexing for searching markdown documents")
	flag.BoolVar(&setting.verbose, "verbose", false,
		"Verbose output")
	flag.BoolVar(&setting.quiet, "quiet", false,
		"Suppress output")
	flag.BoolVar(&setting.color, "color", true,
		"Colored output")

	// usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "syaro %s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: %s [<flags>] [<wikiroot>]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArgs:\n  wikiroot=\"./\": Root folder of wiki\n")
	}
}

func parseFlags() {
	flag.Parse()

	if len(flag.Args()) > 0 {
		setting.wikiRoot = filepath.Clean(flag.Arg(0)) // set wikiroot
	} else {
		setting.wikiRoot = "."
	}
}

func loadYaml() {
	// default values
	setting.Markdown.MathJax = true
	setting.Markdown.Highlight = true
	setting.Markdown.Emoji = true

	b, err := ioutil.ReadFile(filepath.Join(setting.wikiRoot, yamlPath))
	if err != nil {
		log.Info(yamlPath + " not found")
		return
	}

	err = yaml.Unmarshal(b, &setting)
	if err != nil {
		log.Error("Failed to load "+yamlPath+": %v", err)
		return
	}
	log.Info("Load settings from " + yamlPath)
}
