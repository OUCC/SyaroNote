// +build release

package main

import (
	"log"
	"os"
)

const (
	DEBUG = false
)

var (
	logger *log.Logger
)

func init() {
	out := os.DevNull
	logger = log.New(out, "syaro: ", log.Lshortfile)
}
