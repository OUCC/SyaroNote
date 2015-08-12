package main

import (
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
		// HTML_SMARTYPANTS_FRACTIONS |
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

	log.Debug("setting up the HTML renderer")
	renderer := HtmlRenderer(htmlFlags, "", "")

	log.Debug("rendering html with blackfriday")
	mdHtml := Markdown(input, renderer, extensions)

	log.Debug("parsing html")
	r := bytes.NewReader(mdHtml) // byte reader
	tree, err := html.Parse(r)
	if err != nil {
		log.Panic(err)
	}

	var treeManip func(*html.Node)
	treeManip = func(n *html.Node) {
		// if text node is in `pre` or `code`
		if n.Type == html.ElementNode && (n.Data == "pre" || n.Data == "code") {
			return
		}

		// search for [[WikiLink]]
		if n.Type == html.TextNode {
			processWikiLink(n, dir)
		}

		// search for [[WikiLink]](Page Name)
		if n.Type == html.ElementNode && n.Data == "a" {
			processWikiLink2(n, dir)
		}

		// task list
		if n.Type == html.ElementNode && (n.Data == "ul" || n.Data == "ol") {
			processTaskList(n)
		}

		// nav
		if n.Type == html.ElementNode && (n.Data == "nav") {
			processNav(n)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			treeManip(c)
		}
	}

	treeManip(tree)

	log.Debug("re-rendering html from tree")
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

func processNav(nav *html.Node) {
	empty := false
	c := nav.FirstChild
	var ul *html.Node
	if c == nil {
		log.Debug("nav has no child")
		empty = true
	} else {
		// search ul
	loop:
		for {
			switch {
			case c == nil:
				log.Debug("nav has no ul")
				empty = true
				break loop
			case c.Type == html.ElementNode && c.Data == "ul":
				log.Debug("nav has ul")
				ul = c
				break loop
			}
			c = c.NextSibling
		}
	}

	if empty {
		log.Debug("remove nav")
		n := nav.Parent
		if n.FirstChild == nav {
			n.FirstChild = nav.NextSibling
		}
		if n.LastChild == nav {
			n.LastChild = nav.PrevSibling
		}
		if nav.PrevSibling != nil {
			nav.PrevSibling.NextSibling = nav.NextSibling
		}
		if nav.NextSibling != nil {
			nav.NextSibling.PrevSibling = nav.PrevSibling
		}
		return
	}

	var process func(*html.Node) *html.Node
	process = func(ul *html.Node) *html.Node {
		if ul.FirstChild == nil {
			log.Debug("ul has no child")
			return nil
		}

		// search li
		itemNum := 0
		c := ul.FirstChild
		var li *html.Node
	loop:
		for {
			switch {
			case c == nil:
				break loop
			case c.Type == html.ElementNode && c.Data == "li":
				itemNum++
				li = c
			}
			c = c.NextSibling
		}

		switch itemNum {
		case 0:
			log.Debug("ul has no li")
			return nil
		case 1:
			log.Debug("ul has only one li")

			if li.FirstChild == nil {
				log.Debug("li has no child")
				return nil
			}

			// search ul
			c = li.FirstChild
			for {
				switch {
				case c == nil:
					log.Debug("li has no ul")
					return nil
				case c.Type == html.ElementNode && c.Data == "ul":
					log.Debug("li has ul")
					return process(c)
				}
				c = c.NextSibling
			}
		default:
			log.Debug("ul has %d li", itemNum)
			return ul
		}
	}

	switch ul = process(ul); ul {
	case nil:
		// do nothing
		log.Debug("do nothing")
		return
	default:
		log.Debug("set ul as only one child of nav")
		ul.NextSibling = nil
		ul.PrevSibling = nil
		nav.FirstChild = ul
		nav.LastChild = ul

		log.Debug("add .toc-toggle before ul")
		toggle := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{
				html.Attribute{Key: "class", Val: "toc-toggle"},
			},
		}
		span := &html.Node{
			Type: html.ElementNode,
			Data: "span",
		}
		label := &html.Node{
			Type: html.TextNode,
			Data: "TOC",
		}
		caret := &html.Node{
			Type: html.ElementNode,
			Data: "i",
			Attr: []html.Attribute{
				html.Attribute{Key: "class", Val: "glyphicon glyphicon-chevron-up"},
			},
		}
		span.FirstChild = label
		span.LastChild = label
		span.NextSibling = caret
		toggle.FirstChild = span
		toggle.LastChild = caret
		toggle.NextSibling = ul
		nav.FirstChild = toggle
		ul.PrevSibling = toggle

		nav.Attr = []html.Attribute{
			html.Attribute{Key: "class", Val: "toc-open"},
		}
	}
}
