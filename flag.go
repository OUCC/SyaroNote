package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	setting Setting
)

type Setting struct {
	syaroDir   string
	wikiName   string
	wikiRoot   string
	port       int
	urlPrefix  string
	fcgi       bool
	mathjax    bool
	highlight  bool
	singleFile bool // TODO
	readonly   bool // TODO
	gitMode    bool
	verbose    bool
	quiet      bool
	color      bool
}

func init() {
	// setting up flags
	flag.StringVar(&setting.wikiName, "wikiname", "Syaro Wiki",
		"Name of wiki")
	flag.IntVar(&setting.port, "port", 8080,
		"Port number")
	flag.StringVar(&setting.urlPrefix, "url-prefix", "",
		"URL prefix (ex. if prefix is syarowiki, URL is localhost:PORT/syarowiki/)")
	flag.BoolVar(&setting.fcgi, "fcgi", false,
		"If true, syaro runs on fast cgi mode")
	flag.BoolVar(&setting.mathjax, "mathjax", true,
		"MathJax (Internet connection is required)")
	flag.BoolVar(&setting.highlight, "highlight", true,
		"Syntax highlighting in <code> (Internet connection is required)")
	flag.BoolVar(&setting.singleFile, "single", false,
		"Single file mode")
	flag.BoolVar(&setting.gitMode, "gitmode", false,
		"Enable git integration")
	flag.BoolVar(&setting.readonly, "readonly", false,
		"Readonly mode")
	flag.BoolVar(&setting.verbose, "verbose", false,
		"Verbose output")
	flag.BoolVar(&setting.quiet, "quiet", false,
		"Suppress output")
	flag.BoolVar(&setting.color, "color", true,
		"Colored output")

	// usage
	flag.Usage = func() {
		// fmt.Fprintf(os.Stderr, "syaro %s\n\n", version)
		fmt.Fprintf(os.Stderr, "syaro %s\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [<flags>] [<wikiroot>]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArgs:\n  wikiroot=\"./\": Root folder of wiki\n")
	}
}

func parseFlags() {
	flag.Parse()

	if len(flag.Args()) > 0 {
		setting.wikiRoot = flag.Arg(0) // set wikiroot
	} else {
		setting.wikiRoot = "."
	}

	// TODO os.Getenv
}
