package main

import (
	. "github.com/OUCC/syaro/logger"

	. "github.com/russross/blackfriday"

	"code.google.com/p/go.net/html"

	"bytes"
)

func parseMarkdown(input []byte, dir string) []byte {
	htmlFlags := 0 |
		HTML_HREF_TARGET_BLANK |
		HTML_TOC |
		HTML_USE_XHTML |
		HTML_USE_SMARTYPANTS |
		HTML_SMARTYPANTS_FRACTIONS |
		HTML_SMARTYPANTS_LATEX_DASHES |
		HTML_FOOTNOTE_RETURN_LINKS

	extensions := 0 |
		EXTENSION_NO_INTRA_EMPHASIS |
		EXTENSION_TABLES |
		EXTENSION_FENCED_CODE |
		EXTENSION_AUTOLINK |
		EXTENSION_STRIKETHROUGH |
		EXTENSION_SPACE_HEADERS |
		EXTENSION_FOOTNOTES |
		EXTENSION_HEADER_IDS |
		EXTENSION_AUTO_HEADER_IDS

	LoggerV.Println("main.parseMarkdown: setting up the HTML renderer")
	renderer := HtmlRenderer(htmlFlags, "", "")

	LoggerV.Println("main.parseMarkdown: rendering html with blackfriday")
	mdHtml := Markdown(input, renderer, extensions)

	LoggerV.Println("main.parseMarkdown: parsing html")
	r := bytes.NewReader(mdHtml) // byte reader
	tree, err := html.Parse(r)
	if err != nil {
		LoggerE.Panicln("main.ParseMarkdown: Error!", err)
	}

	var treeManip func(*html.Node)
	treeManip = func(n *html.Node) {
		// if text node is in `pre` of `code`
		if n.Type == html.ElementNode && (n.Data == "pre" || n.Data == "code") {
			return
		}

		if n.Type == html.TextNode {
			// search for [[WikiLink]]
			processWikiLink(n, dir)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			treeManip(c)
		}
	}

	treeManip(tree)

	LoggerV.Println("main.ParseMarkdown: re-rendering html from tree")
	var w bytes.Buffer
	html.Render(&w, tree) // re-render html
	b := w.Bytes()

	// strip prefix (<html><head></head><body>) and suffix (</body></html>)
	return b[25 : len(b)-14]
}
