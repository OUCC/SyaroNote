package main

import (
	"github.com/op/go-logging"

	"os"
)

const (
	LOG_FORMAT       = "%{time:2006/01/02 15:04:05.000000} syaro %{shortfunc:-12.12s} | %{level:.4s} %{message}"
	COLOR_LOG_FORMAT = "%{time:2006/01/02 15:04:05.000000} syaro %{shortfunc:-12.12s} | %{color:bold}%{level:.4s}%{color:reset} %{color}%{message}%{color:reset}"
)

var (
	log = logging.MustGetLogger("syaro")
)

func setupLogger() {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	var format logging.Formatter
	if setting.color {
		format = logging.MustStringFormatter(COLOR_LOG_FORMAT)
	} else {
		format = logging.MustStringFormatter(LOG_FORMAT)
	}
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	if setting.verbose {
		backendLeveled.SetLevel(logging.DEBUG, "")
	} else if setting.quiet {
		backendLeveled.SetLevel(logging.ERROR, "")
	} else {
		backendLeveled.SetLevel(logging.INFO, "")
	}
	log.SetBackend(backendLeveled)
}
