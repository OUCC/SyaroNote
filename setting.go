package main

import (
	"flag"
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
	VERBOSE_ARGNAME  = "verbose"
	VERBOSE_ARGNAMES = "v"
	VERBOSE_USAGE    = "Verbose output."
	VERBOSE_DEFAULT  = false
)

// TODO wikiname
type Setting struct {
	wikiRoot  string
	syaroDir  string
	port      int
	urlPrefix string
	fcgi      bool
	verbose   bool
}

func init() {
	setting = new(Setting)

	flag.StringVar(&setting.wikiRoot, WIKIROOT_ARGNAME, WIKIROOT_DEFAULT, WIKIROOT_USAGE)
	flag.StringVar(&setting.syaroDir, SYARO_DIR_ARGNAME, SYARO_DIR_DEFAULT, SYARO_DIR_USAGE)
	flag.IntVar(&setting.port, PORT_ARGNAME, PORT_DEFAULT, PORT_USAGE)
	flag.StringVar(&setting.urlPrefix, URL_PREFIX_ARGNAME, URL_PREFIX_DEFAULT, URL_PREFIX_USAGE)
	flag.BoolVar(&setting.fcgi, FCGI_ARGNAME, FCGI_DEFAULT, FCGI_USAGE)
	flag.BoolVar(&setting.verbose, VERBOSE_ARGNAME, VERBOSE_DEFAULT, VERBOSE_USAGE)
	flag.BoolVar(&setting.verbose, VERBOSE_ARGNAMES, VERBOSE_DEFAULT, VERBOSE_USAGE)
}
