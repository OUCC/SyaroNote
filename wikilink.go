package main

import (
	. "github.com/OUCC/syaro/logger"
	"github.com/OUCC/syaro/setting"
	"github.com/OUCC/syaro/util"
	"github.com/OUCC/syaro/wikiio"

	"bufio"
	"bytes"
	"html"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

func processWikiLink(s string, currentDir string) string {
	const RE_DOUBLE_BRACKET = "\\[\\[[^\\]]+\\]\\]"

	reader := strings.NewReader(s)
	scanner := bufio.NewScanner(reader)
	var buffer bytes.Buffer

	re := regexp.MustCompile(RE_DOUBLE_BRACKET)

	for scanner.Scan() {
		line := scanner.Text()

		for {
			indices := re.FindStringIndex(line)

			if len(indices) != 0 { // tag found
				LoggerV.Println("processWikiLink: bracket tag found:",
					string(line[indices[0]:indices[1]]))

				name := line[indices[0]+2 : indices[1]-2] // [[name]]
				files := searchPage(name, currentDir)

				if len(files) != 0 { // page found
					LoggerV.Println("processWikiLink:", len(files), "pages found")
					LoggerV.Println("processWikiLink: select ", files[0].WikiPath())
					// TODO avoid ambiguous page
					line = embedLinkTag(line, indices, name, files[0])

				} else { // page not found
					LoggerV.Println("processWikiLink: no page found")
					line = embedLinkTag(line, indices, name, nil)
				}

			} else { // tag not found, so go next line
				break
			}
		}
		buffer.Write([]byte(line))
		buffer.Write([]byte("\n"))
	}

	return buffer.String()
}

func embedLinkTag(line string, tagIndex []int, linkname string, file *wikiio.WikiFile) string {
	if file == nil {
		return strings.Join([]string{
			line[:tagIndex[0]],
			"<a class=\"notfound\" href=\"",
			setting.UrlPrefix,
			"/error/404?data=",
			url.QueryEscape(string(linkname)),
			"\">",
			linkname,
			"</a>",
			line[tagIndex[1]:],
		}, "")
	}
	return strings.Join([]string{
		line[:tagIndex[0]],
		"<a href=\"",
		string(file.URLPath()),
		"\">",
		linkname,
		"</a>",
		line[tagIndex[1]:],
	}, "")
}

func searchPage(name string, currentDir string) []*wikiio.WikiFile {
	if name == "" {
		return nil
	}

	// unescape for searching
	name = html.UnescapeString(name)

	if filepath.IsAbs(name) {
		// search name as absolute path
		// example: /piyo /poyo/pyon.ext
		return searchPageByAbsPath(name)
	} else if strings.Contains(name, "/") || util.IsMarkdown(name) {
		// search name as relative path
		// example: ./hoge ../fuga.ext puyo.ext
		return searchPageByRelPath(name, currentDir)
	} else {
		// search name as base name
		// example: abc
		return searchPageByBaseName(name)
	}
}

func searchPageByAbsPath(abspath string) []*wikiio.WikiFile {
	LoggerV.Printf("main.searchPageByAbsPath(%s)", abspath)
	file, _ := wikiio.Load(abspath)
	if file == nil {
		return nil
	}
	return []*wikiio.WikiFile{file}
}

func searchPageByRelPath(relpath, currentDir string) []*wikiio.WikiFile {
	LoggerV.Printf("main.searchPageByRelPath(%s, %s)", relpath, currentDir)
	wpath := filepath.Join(currentDir, relpath)
	file, _ := wikiio.Load(wpath)
	if file == nil {
		return nil
	}
	return []*wikiio.WikiFile{file}
}

func searchPageByBaseName(baseName string) []*wikiio.WikiFile {
	LoggerV.Printf("main.searchPageByBaseName(%s)", baseName)
	files, _ := wikiio.Search(baseName)
	return files
}
