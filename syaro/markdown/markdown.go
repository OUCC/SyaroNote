package markdown

import (
	. "github.com/yuntan/blackfriday"

	"code.google.com/p/go.net/html"
	"gopkg.in/yaml.v2"

	"bytes"
)

var LinkWorker func(*bytes.Buffer, []byte)

func processTree(b []byte) []byte {
	var treeManip func(*html.Node)
	treeManip = func(n *html.Node) {
		// if text node is in `pre` or `code`
		if n.Type == html.ElementNode && (n.Data == "pre" || n.Data == "code") {
			return
		}

		// search for [[WikiLink]]
		if n.Type == html.TextNode {
			processEmoji(n)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			treeManip(c)
		}
	}

	r := bytes.NewReader(b) // byte reader
	tree, err := html.Parse(r)
	if err != nil {
		return nil
	}

	treeManip(tree)

	var w bytes.Buffer
	html.Render(&w, tree) // re-render html

	// strip prefix (<html><head></head><body>) and suffix (</body></html>)
	return w.Bytes()[25 : len(b)-14]
}

func Convert(input []byte) []byte {
	if input == nil {
		return nil
	}

	// remove front matter
	sep := []byte("---\n")
	if bytes.HasPrefix(input, sep) {
		b := bytes.SplitN(input, sep, 3)
		if len(b) == 3 {
			input = b[2]
		}
	}

	htmlFlags := 0 |
		HTML_HREF_TARGET_BLANK |
		HTML_USE_XHTML |
		HTML_USE_SMARTYPANTS |
		// HTML_SMARTYPANTS_FRACTIONS |
		HTML_SMARTYPANTS_LATEX_DASHES |
		HTML_FOOTNOTE_RETURN_LINKS |
		HTML_IMAGES_AS_FIGURE

	extensions := 0 |
		EXTENSION_NO_INTRA_EMPHASIS |
		EXTENSION_TABLES |
		EXTENSION_FENCED_CODE |
		EXTENSION_AUTOLINK |
		EXTENSION_STRIKETHROUGH |
		EXTENSION_SPACE_HEADERS |
		EXTENSION_FOOTNOTES |
		EXTENSION_HEADER_IDS |
		EXTENSION_AUTO_HEADER_IDS |
		EXTENSION_BACKSLASH_LINE_BREAK |
		EXTENSION_DEFINITION_LISTS |
		EXTENSION_WIKI_LINK |
		EXTENSION_LATEX_MATH

	renderer := HtmlRenderer(htmlFlags, "", "")
	if LinkWorker != nil {
		WikiLinkWorker = LinkWorker
	}
	return processTree(Markdown(input, renderer, extensions))
}

func Meta(input []byte) map[string]string {
	// get front matter
	sep := []byte("---\n")
	if !bytes.HasPrefix(input, sep) {
		return nil
	}
	b := bytes.SplitN(input, sep, 3)
	if len(b) != 3 {
		return nil
	}

	ret := make(map[string]string)
	// parse front matter
	yaml.Unmarshal(b[1], &ret)
	return ret
}

func TOC(input []byte) []byte {
	// remove front matter
	sep := []byte("---\n")
	if bytes.HasPrefix(input, sep) {
		b := bytes.SplitN(input, sep, 3)
		if len(b) == 3 {
			input = b[2]
		}
	}

	htmlFlags := 0 |
		HTML_TOC |
		HTML_OMIT_CONTENTS |
		HTML_USE_XHTML

	extensions := 0 |
		EXTENSION_SPACE_HEADERS |
		EXTENSION_HEADER_IDS |
		EXTENSION_AUTO_HEADER_IDS

	return processTree(Markdown(input, HtmlRenderer(htmlFlags, "", ""), extensions))
}
