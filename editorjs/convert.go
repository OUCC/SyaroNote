package main

import (
	"github.com/OUCC/syaro/markdown"
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	js.Global.Set("convert", func(text string) string {
		return string(markdown.Convert([]byte(text)))
	})
}
