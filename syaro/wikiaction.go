package main

import (
	"os/exec"
	"strings"
)

const (
	ACTION_CREATE = 1 + iota
	ACTION_UPDATE
	ACTION_RENAME
	ACTION_DELETE
	ACTION_UPLOAD
)

type wikiAction struct {
	action   int
	wikiPath string
	message  string
	name     string
	email    string
}

func doPostAction(wa wikiAction) {
	var ss []string
	switch wa.action {
	case ACTION_CREATE:
		ss = setting.Actions.PostCreate
	case ACTION_UPDATE:
		ss = setting.Actions.PostUpdate
	case ACTION_RENAME:
		ss = setting.Actions.PostRename
	case ACTION_DELETE:
		ss = setting.Actions.PostUpload
	case ACTION_UPLOAD:
		ss = setting.Actions.PostUpload
	}
	if len(ss) == 0 {
		return
	}

	for i, s := range ss {
		s = strings.Replace(s, "{{wikiPath}}", wa.wikiPath, -1)
		s = strings.Replace(s, "{{message}}", wa.message, -1)
		s = strings.Replace(s, "{{name}}", wa.name, -1)
		s = strings.Replace(s, "{{email}}", wa.email, -1)
		ss[i] = s
	}

	log.Info("Executing action `%s`...", strings.Join(ss, " "))
	cmd := exec.Command(ss[0], ss[1:len(ss)]...)
	if err := cmd.Start(); err != nil {
		log.Error("Failed to start action: %v", err)
		return
	}
	log.Info("Done")
}
