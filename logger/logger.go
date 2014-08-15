package logger

import (
	"github.com/OUCC/syaro/setting"

	"io/ioutil"
	"log"
	"os"
)

var (
	LoggerM *log.Logger
	LoggerV *log.Logger
	LoggerE *log.Logger
)

func SetupLogger() {
	prefix := "syaro: "
	flag := log.Ldate | log.Ltime | log.Lshortfile

	LoggerM = log.New(os.Stdout, prefix, flag)

	if setting.Verbose {
		LoggerV = log.New(os.Stdout, prefix, flag)
	} else {
		LoggerV = log.New(ioutil.Discard, prefix, flag)
	}

	LoggerE = log.New(os.Stderr, prefix, flag)
}
