package main

import (
	pb "github.com/OUCC/syaro/gitservice"

	"net/http"
)

const (
	HISTORY_VIEW = "history.html"
)

type HistoryPage struct {
	*Page
	// Changes []Change
	Changes []*pb.Change
}

func LoadHistoryPage(wpath string) (*HistoryPage, error) {
	log.Debug("wpath: %s", wpath)

	page, err := LoadPage(wpath)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Debug("ok")
	return &HistoryPage{
		Page:    page,
		Changes: page.History(),
	}, nil
}

func (page *HistoryPage) Title() string {
	return "History of " + page.NameWithoutExt()
}

func (page *HistoryPage) Render(res http.ResponseWriter) error {
	return tmpl.ExecuteTemplate(res, HISTORY_VIEW, page)
}
