package main

import (
	"github.com/op/go-logging"

	"os"
)

const (
	LOG_FORMAT = "%{time:2006/01/02 15:04:05.000000} gitpl %{shortfunc:-12.12s} | %{level:.4s} %{message}"
)

var (
	log = logging.MustGetLogger("gitplugin")
)

func setupLogger() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	format := logging.MustStringFormatter(LOG_FORMAT)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(logging.DEBUG, "")
	log.SetBackend(backendLeveled)
}
