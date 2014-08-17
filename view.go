package main

import (
	"net/http"
)

type View interface {
	Render(http.ResponseWriter) error
}
