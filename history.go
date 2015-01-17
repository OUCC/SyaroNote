package main

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/wikiio"

	"net/http"
)

const (
	HISTORY_VIEW = "history.html"
)

type HistoryPage struct {
	*Page
	Changes []wikiio.Change
}

func LoadHistoryPage(wpath string) (*HistoryPage, error) {
	Log.Debug("wpath: %s", wpath)

	page, err := LoadPage(wpath)
	if err != nil {
		Log.Error(err.Error())
		return nil, err
	}

	Log.Debug("ok")
	return &HistoryPage{
		Page:    page,
		Changes: page.History(),
	}, nil
}

func (page *HistoryPage) Title() string {
	return "History of " + page.NameWithoutExt()
}

func (page *HistoryPage) Render(res http.ResponseWriter) error {
	return views.ExecuteTemplate(res, HISTORY_VIEW, page)
}
