Syaro Markdown Wiki Server
====

[![Build Status](https://drone.io/github.com/OUCC/syaro/status.png)](https://drone.io/github.com/OUCC/syaro/latest)

Description
----
Syaro is a simple and pretty wiki system.

Features
----
### Markdown Viewer
Syaro can handle markdown format files. [marked] is used to convert
Markdown to HTML.

Viewer supports MathJax. LaTeX text surrounded by `$` (inline math) or `$$` 
(display math) is converted to beautiful mathematical expression.

Viewer also supports code syntax highlighting. This feature is powered by
[highlight.js].

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

### Powerful Markdown editor
* Realtime preview (including MathJax rendering and code highlighting)
* Markdown syntax highlight

Build & Install
----
First, install go and bower.

```bash
go install github.com/OUCC/syaro.git
cd $GOPATH/src/github.com/OUCC/syaro
# get jquery, bootstrap, etc...
bower install
./install_components.sh # copy files
```

Use
---
```bash
syaro --wikiroot=/path/to/wiki

# If you want to use MathJax or highlight.js,
# syaro --mathjax --highlight --wikiroot=...
```

Then open `localhost:8080/Home` in your browser.

`syaro -h` or `syaro --help` you can see more options.

About this software
----
This software is released under MIT License.

Following software is used:

* [Go]  (BSD)
* [Twitter Bootstrap]  (MIT)
* [jQuery]  (MIT)
* [Ace]  (BSD)
* [marked]  (MIT)
* [MathJax]  (Apache)
* [highlight.js]  (BSD)
* ([dillinger])  (MIT)


[Go]: http://golang.org/
[Twitter Bootstrap]: http://getbootstrap.com
[jQuery]: http://jquery.com
[Ace]: http://ace.c9.io
[marked]: https://github.com/chjj/marked
[Mathjax]: http://www.mathjax.org/
[highlight.js]: https://highlightjs.org/
[dillinger]: https://github.com/joemccann/dillinger/
