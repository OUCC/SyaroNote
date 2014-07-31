// +build !release

package main

import (
	"log"
	"os"
)

const (
	DEBUG = true
)

var (
	logger *log.Logger
)

func init() {
	out := os.Stdout
	logger = log.New(out, "syaro: ", log.Lshortfile)
}
