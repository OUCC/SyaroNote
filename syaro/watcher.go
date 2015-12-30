package main

import (
	"gopkg.in/fsnotify.v1"

	"os"
	"path/filepath"
	"strings"
)

func fsWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Failed to setting up filesystem watcher: %v", err)
		log.Error("Auto reload and text indexing will not be available")
		return
	}
	defer watcher.Close()

	log.Info("Adding files to watcher...")
	filepath.Walk(setting.wikiRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error(err.Error())
			return nil
		}

		// exclude hidden dir (ex. .git)
		if info.IsDir() && !strings.Contains(path, "/.") {
			watcher.Add(path)
			log.Debug("%s added to watcher", path)
		}

		return nil
	})

	// event loop for watcher
	for {
		select {
		case event := <-watcher.Events:
			log.Debug("%+v", event)
			updateIndex <- event

		case err := <-watcher.Errors:
			log.Error("Filesystem watcher unexpectedly crashed: %v", err)
		}
	}
}
