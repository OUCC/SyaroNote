package main

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	loggerM *log.Logger
	loggerV *log.Logger
	loggerE *log.Logger
)

func setupLogger() {
	prefix := "syaro: "
	flag := log.Ldate | log.Ltime | log.Lshortfile

	loggerM = log.New(os.Stdout, prefix, flag)

	if setting.verbose {
		loggerV = log.New(os.Stdout, prefix, flag)
	} else {
		loggerV = log.New(ioutil.Discard, prefix, flag)
	}

	loggerE = log.New(os.Stderr, prefix, flag)
}
