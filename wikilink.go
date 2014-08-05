package main

import (
	"bufio"
	"bytes"
	"container/list"
	"os"
	"path/filepath"
	"regexp"
)

type WikiLink struct {
	Name     []byte
	FilePath []byte
	Title    []byte
}

func processWikiLink(b []byte, currentDir string) []byte {
	const reDoubleBracket = "\\[\\[[^\\]]+\\]\\]"

	reader := bytes.NewReader(b)
	scanner := bufio.NewScanner(reader)
	var buffer bytes.Buffer

	re := regexp.MustCompile(reDoubleBracket)

	for scanner.Scan() {
		line := scanner.Bytes()

		for {
			index := re.FindIndex(line)

			if len(index) != 0 { // tag found
				loggerV.Println("bracket tag found:", string(line[index[0]:index[1]]))

				name := line[index[0]+2 : index[1]-2]
				links, err := searchPage(string(name), currentDir)
				if err != nil {
					loggerE.Fatalf("error occurd in serarchPage(%s, %s)", string(name), currentDir)
					loggerE.Fatalln(err.Error())
					continue
				}

				if len(links) != 0 { // page found
					// TODO avoid ambiguous page
					line = embedLinkTag(line, index, links[0])

				} else { // page not found
					// TODO invalid link
					line = embedLinkTag(line, index, WikiLink{Name: name})
				}

			} else { // tag not found, so go next line
				break
			}
		}
		buffer.Write(line)
	}

	return buffer.Bytes()
}

func embedLinkTag(line []byte, tagIndex []int, link WikiLink) []byte {
	return bytes.Join([][]byte{
		line[:tagIndex[0]],
		[]byte("<a href=\""),
		[]byte(filepath.Join("/", setting.urlPrefix, string(link.FilePath))),
		[]byte("\" title=\""),
		link.Title,
		[]byte("\">"),
		link.Name,
		[]byte("</a>"),
		line[tagIndex[1]:]}, nil)
}

// TODO security check
func searchPage(name string, currentDir string) ([]WikiLink, error) {
	if name == "" {
		return nil, nil
	}

	// TODO
	// if filepath.IsAbs(name) {
	// search name as absolute path
	// example: /piyo /poyo/pyon.ext
	// paths, err := searchPageByAbsPath(name, currentDir)
	// } else if strings.Contains(name, "/") || isMarkdown(name) {
	// search name as relative path
	// example: ./hoge ../fuga.ext puyo.ext
	// paths, err := searchPageByRelPath(name, currentDir)
	// } else {
	// search name as base name
	// example: abc
	paths, err := searchPageByBaseName(name)
	// }

	if err != nil {
		return nil, err
	}

	if len(paths) != 0 {
		ret := make([]WikiLink, len(paths))
		for i, path := range paths {
			ret[i] = WikiLink{
				Name:     []byte(name),
				FilePath: []byte(path),
				Title:    nil, // TODO
			}
		}
		return ret, nil
	}

	// not found
	return nil, nil
}

func searchPageByBaseName(baseName string) ([]string, error) {
	loggerV.Printf("searchPageByBaseName(%s)", baseName)

	foundPath := list.New()

	// func for filepath.Walk
	// This judges whether path match to baseName, and add path to list
	walkfunc := func(path string, info os.FileInfo, err error) error {
		path = "/" + path // make it wiki path

		if removeExt(info.Name()) == baseName {
			if info.IsDir() {
				loggerV.Println("dir found!", path)
				foundPath.PushBack(path)

			} else {
				loggerV.Println("page found!", path)

				if filepath.Base(filepath.Dir(path)) == baseName {
					loggerV.Print("this page is main page of dir %s. Ignored.",
						filepath.Dir(path))
				} else {
					foundPath.PushBack(path)
				}
			}
		}
		return nil
	}

	err := filepath.Walk(setting.wikiRoot, walkfunc)
	if err != nil {
		return nil, err
	}

	loggerV.Println(foundPath.Len(), "pages found")

	return toStringArray(foundPath), nil
}
