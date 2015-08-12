package main

import (
	"gopkg.in/fsnotify.v1"

	"os"
	"path/filepath"
	"strings"
)

var (
	// file system watcher
	watcher *fsnotify.Watcher
)

func initWatcher() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// event loop for watcher
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Debug("%s", event)
				switch {
				case event.Op&fsnotify.Create != 0:
					log.Info("New file Created (%s)", event.Name)
					refreshRequired = true

				case event.Op&fsnotify.Remove != 0:
					log.Info("File removed (%s)", event.Name)
					refreshRequired = true
				}

			case err := <-watcher.Errors:
				log.Fatal(err)
			}
		}
	}()

	filepath.Walk(setting.wikiRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error(err.Error())
		}

		// dont add hidden dir (ex. .git) and backup file
		if info.IsDir() &&
			!strings.Contains(path, "/.") &&
			!strings.HasPrefix(path, ".") &&
			!strings.HasSuffix(path, BACKUP_SUFFIX) {
			watcher.Add(path)
			log.Debug("%s added to watcher", path)
		}

		return nil
	})
}

func closeWatcher() {
	watcher.Close()
}
