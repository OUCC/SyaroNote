package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	WIKINAME_ARGNAME = "wikiname"
	WIKINAME_USAGE   = "Name of wiki."
	WIKINAME_DEFAULT = "Syaro Wiki"

	PORT_ARGNAME = "port"
	PORT_USAGE   = "Port."
	PORT_DEFAULT = 8080

	URL_PREFIX_ARGNAME = "url-prefix"
	URL_PREFIX_USAGE   = "URL prefix (ex. if prefix is syarowiki, URL is localhost:PORT/syarowiki/)."
	URL_PREFIX_DEFAULT = ""

	FCGI_ARGNAME = "fcgi"
	FCGI_USAGE   = "If true, syaro runs on fast cgi mode."
	FCGI_DEFAULT = false

	// TODO
	MATHJAX_ARGNAME = "no-mathjax"
	MATHJAX_USAGE   = "Disable MathJax (Internet connection is required)"
	MATHJAX_DEFAULT = false

	// TODO
	HIGHLIGHT_ARGNAME = "no-highlight"
	HIGHLIGHT_USAGE   = "Disable syntax highlighting in <code> (Internet connection is required)"
	HIGHLIGHT_DEFAULT = false

	// TODO
	READONLY_ARGNAME = "readonly"
	READONLY_USAGE   = "Enable readonly mode."
	READONLY_DEFAULT = false

	VERBOSE_ARGNAME  = "verbose"
	VERBOSE_ARGNAMES = "v"
	VERBOSE_USAGE    = "Verbose output."
	VERBOSE_DEFAULT  = false
)

var (
	setting Setting
)

type Setting struct {
	wikiName string
	wikiRoot string
	port int
	urlPrefix string
	fcgi      bool
	mathjax bool
	highlight bool
	verbose bool
	gitMode bool
}

func init() {
	// setting up flags
	flag.StringVar(&setting.wikiName, WIKINAME_ARGNAME, WIKINAME_DEFAULT, WIKINAME_USAGE)
	flag.IntVar(&setting.port, PORT_ARGNAME, PORT_DEFAULT, PORT_USAGE)
	flag.StringVar(&setting.urlPrefix, URL_PREFIX_ARGNAME, URL_PREFIX_DEFAULT, URL_PREFIX_USAGE)
	flag.BoolVar(&setting.fcgi, FCGI_ARGNAME, FCGI_DEFAULT, FCGI_USAGE)
	//flag.BoolVar(&setting.mathjax, MATHJAX_ARGNAME, MATHJAX_DEFAULT, MATHJAX_USAGE)
	//flag.BoolVar(&setting.highlight, HIGHLIGHT_ARGNAME, HIGHLIGHT_DEFAULT, HIGHLIGHT_USAGE)
	flag.BoolVar(&setting.verbose, VERBOSE_ARGNAME, VERBOSE_DEFAULT, VERBOSE_USAGE)
	flag.BoolVar(&setting.verbose, VERBOSE_ARGNAMES, VERBOSE_DEFAULT, VERBOSE_USAGE)

	// usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "syaro %s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: %s [<flags>] [<wikiroot>]\n\nFlags:", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArgs:\n  wikiroot=\"./\": Root folder of wiki")
	}
}

func parseFlags() {
	flag.Parse()

	setting.wikiRoot = flag.Arg(0) // set wikiroot

	// TODO os.Getenv
}
