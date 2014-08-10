Syaro Markdown Wiki Server
====

## Description
Syaro is a simple and pretty wiki system.

## Features
### Markdown Viewer
Syaro can handle markdown format files.

### [[WikiLink]]
Texts surrounded by double bracket are interpreted as WikiLink. To link to
another wiki page, you can use WikiLink. For example,

```
[[My Profile Page]]
```

is a link to `My Profile Page.[md,mkd,markdown]`. Syaro searches all files under
wikiroot, so two files haven't be in same folder.

### Files/Folders under ...
In main page of a directory, file/folder list is automatically appended to
page's end.

Example file tree

```
WIKIROOT/ - Home.md
            kinmoza.md
            gochiusa/ - gochiusa.md
                        cocoa.md
                        chino.md
                        rize.md
                        syaro.md
                        chiya.md
```

In this example, main page of `gochiusa/` is `gochiusa.md`, and file/folder list
is appended when you see `gochiusa.md`.

## Install
```bash
git clone https://github.com/OUCC/syaro.git
cd syaro
go install
```

## Use
`cd` to wiki root directory, then run `syaro`. Open `localhost:8080/Home` in
your browser.

`syaro -h` or `syaro --help` you can see options.

## About this software
This software is distributed under MIT License.

Following software is used:
* [Go](http://golang.org/)
* [Blackfriday: a markdown processor for Go](https://github.com/russross/blackfriday)
