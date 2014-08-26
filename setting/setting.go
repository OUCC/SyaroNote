package setting

import (
	"flag"
)

const (
	WIKINAME_ARGNAME = "wikiname"
	WIKINAME_USAGE   = "Name of wiki."
	WIKINAME_DEFAULT = "Syaro Wiki"
)

const (
	WIKIROOT_ARGNAME = "wikiroot"
	WIKIROOT_USAGE   = "Root folder of wiki."
	WIKIROOT_DEFAULT = "./"
)

const (
	SYARO_DIR_ARGNAME = "syaro-dir"
	SYARO_DIR_USAGE   = "Directory for HTML, css, js etc."
	SYARO_DIR_DEFAULT = ""
)

const (
	PORT_ARGNAME = "port"
	PORT_USAGE   = "Port."
	PORT_DEFAULT = 8080
)

const (
	URL_PREFIX_ARGNAME = "url-prefix"
	URL_PREFIX_USAGE   = "URL prefix (ex. if prefix is syarowiki, URL is localhost:PORT/syarowiki/)."
	URL_PREFIX_DEFAULT = ""
)

const (
	FCGI_ARGNAME = "fcgi"
	FCGI_USAGE   = "If true, syaro runs on fast cgi mode."
	FCGI_DEFAULT = false
)

const (
	MATHJAX_ARGNAME = "mathjax"
	MATHJAX_USAGE   = "Enable MathJax (Internet connection is required)"
	MATHJAX_DEFAULT = false
)

const (
	HIGHLIGHT_ARGNAME = "highlight"
	HIGHLIGHT_USAGE   = "Enable syntax highlighting in <code> (Internet connection is required)"
	HIGHLIGHT_DEFAULT = false
)

// TODO
const (
	READONLY_ARGNAME = "readonly"
	READONLY_USAGE   = "Enable readonly mode."
	READONLY_DEFAULT = false
)

const (
	VERBOSE_ARGNAME  = "verbose"
	VERBOSE_ARGNAMES = "v"
	VERBOSE_USAGE    = "Verbose output."
	VERBOSE_DEFAULT  = false
)

// TODO wikiname
var (
	WikiName  string
	WikiRoot  string
	SyaroDir  string
	Port      int
	UrlPrefix string
	FCGI      bool
	MathJax   bool
	Highlight bool
	Verbose   bool
)

func init() {
	flag.StringVar(&WikiName, WIKINAME_ARGNAME, WIKINAME_DEFAULT, WIKINAME_USAGE)
	flag.StringVar(&WikiRoot, WIKIROOT_ARGNAME, WIKIROOT_DEFAULT, WIKIROOT_USAGE)
	flag.StringVar(&SyaroDir, SYARO_DIR_ARGNAME, SYARO_DIR_DEFAULT, SYARO_DIR_USAGE)
	flag.IntVar(&Port, PORT_ARGNAME, PORT_DEFAULT, PORT_USAGE)
	flag.StringVar(&UrlPrefix, URL_PREFIX_ARGNAME, URL_PREFIX_DEFAULT, URL_PREFIX_USAGE)
	flag.BoolVar(&FCGI, FCGI_ARGNAME, FCGI_DEFAULT, FCGI_USAGE)
	flag.BoolVar(&MathJax, MATHJAX_ARGNAME, MATHJAX_DEFAULT, MATHJAX_USAGE)
	flag.BoolVar(&Highlight, HIGHLIGHT_ARGNAME, HIGHLIGHT_DEFAULT, HIGHLIGHT_USAGE)
	flag.BoolVar(&Verbose, VERBOSE_ARGNAME, VERBOSE_DEFAULT, VERBOSE_USAGE)
	flag.BoolVar(&Verbose, VERBOSE_ARGNAMES, VERBOSE_DEFAULT, VERBOSE_USAGE)
}
