package logger

import (
	"github.com/OUCC/syaro/setting"

	logging "github.com/op/go-logging"

	"os"
)

var (
	Log    = logging.MustGetLogger("syaro")
	format = logging.MustStringFormatter(
		"%{time:2006/01/02 15:04:05} %{shortpkg:-6.6s} %{shortfunc:-12.12s} | %{color:bold}%{level:.4s}%{color:reset} %{color}%{message}%{color:reset}",
	)
)

func SetupLogger() {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	if setting.Verbose {
		backendLeveled.SetLevel(logging.DEBUG, "")
	} else {
		backendLeveled.SetLevel(logging.INFO, "")
	}
	Log.SetBackend(backendLeveled)
}
