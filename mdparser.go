package main

import (
	. "github.com/OUCC/syaro/logger"

	. "github.com/russross/blackfriday"

	"code.google.com/p/go.net/html"

	"bytes"
	"regexp"
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
		// if text node is in `pre` or `code`
		if n.Type == html.ElementNode && (n.Data == "pre" || n.Data == "code") {
			return
		}

		if n.Type == html.TextNode {
			// search for [[WikiLink]]
			processWikiLink(n, dir)
		}

		// task list
		if n.Type == html.ElementNode && (n.Data == "ul" || n.Data == "ol") {
			processTaskList(n)
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

// * [ ] checkbox1
// * [x] checkbox2
//
// n is ol or ul ElementNode
func processTaskList(n *html.Node) {
	const RE_CHECKBOX = "^\\[ \\]"
	const RE_CHECKBOX_CHECKED = "^\\[x\\]"

	// n is not ul nor li
	if n.Type != html.ElementNode || (n.Data != "ul" && n.Data != "ol") {
		return
	}

	re := regexp.MustCompile(RE_CHECKBOX)
	reChecked := regexp.MustCompile(RE_CHECKBOX_CHECKED)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "li" {
			continue
		}

		for c2 := c.FirstChild; c2 != nil; c2 = c2.NextSibling {
			if c2.Type == html.TextNode && (re.MatchString(c2.Data) || reChecked.MatchString(c2.Data)) {
				// <ul>
				//   <li class="task-list-item">
				//     <input type="checkbox" class="task-list-item-checkbox" checked disabled>
				//     task1
				//   </li>
				// </ul>
				checked := false
				if reChecked.MatchString(c2.Data) {
					checked = true
				}

				found := false
				for _, attr := range c.Attr {
					if attr.Key == "class" {
						attr.Val += " task-list-item"
						found = true
					}
				}
				if !found {
					c.Attr = append(c.Attr, html.Attribute{
						Key: "class",
						Val: "task-list-item",
					})
				}

				inputElem := html.Node{
					Type: html.ElementNode,
					Data: "input",
					Attr: []html.Attribute{
						html.Attribute{
							Key: "type",
							Val: "checkbox",
						},
						html.Attribute{
							Key: "class",
							Val: "task-list-item-checkbox",
						},
						html.Attribute{
							Key: "disabled",
						},
					},
				}

				if checked {
					inputElem.Attr = append(inputElem.Attr, html.Attribute{
						Key: "checked",
					})
				}

				c.InsertBefore(&inputElem, c2)

				// remove "[ ]" or "[x]"
				if checked {
					c2.Data = reChecked.ReplaceAllString(c2.Data, "")
				} else {
					c2.Data = re.ReplaceAllString(c2.Data, "")
				}
			}

			break
		}
	}
}
