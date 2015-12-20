package markdown

import (
	. "github.com/yuntan/blackfriday"
	"gopkg.in/yaml.v2"

	"bytes"
)

var LinkWorker func(*bytes.Buffer, []byte)

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
	return Markdown(input, renderer, extensions)
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

	return Markdown(input, HtmlRenderer(htmlFlags, "", ""), extensions)
}
