# [SyaroNote](https://syaro.untan.xyz/_.md)
SyaroNote is Markdown wiki for personal and small group.

[Features](https://syaro.untan.xyz/Features)

## Build
Go and nodejs are required.

### syaro
```sh
go get github.com/OUCC/SyaroNote
go build -ldflags "-X main.version=$(git describe)" -o build/syaro github.com/OUCC/SyaroNote/syaro
```

### editor
```sh
cd editorjs
npm install
npm run deploy
```

### copy public
```sh
npm install
bower install
gulp copy
```

### git plugin (optional)
```sh
cd $GOPATH/src/github.com/libgit2/git2go
git submodule update --init
make install
go install

cd $GOPATH/src/github.com/OUCC/SyaroNote
go build -o build/gitplugin github.com/OUCC/SyaroNote/gitplugin
```

Usage
----
```bash
./syaro path/to/wiki

# if public and template folders placed outer than working dir
SYARODIR=/path/to/syaro/dir syaro path/to/wiki
```

Then open `localhost:8080/Home` in your browser.

`syaro -h` or `syaro --help` you can see more options.

Contribution
----
Fork and pull requests welcome.

Donate
----
* [Donate $3](https://gumroad.com/l/Jwtx)
* [誰かに買って欲しいものリスト](http://www.amazon.co.jp/registry/wishlist/1MVMC2QBIJYY)

About
----
### Author
* @yuntan
* @susisu (Table editor)
* @spring-raining (emoji)

### License
This software is released under MIT License.

### Softwares
Following softwares are used:

* [Go]  (BSD)
* [blackfriday]  (BSD)
* [go-logging]
* [Twitter Bootstrap]  (MIT)
* [jQuery]  (MIT)
* [Ace]  (BSD)
* [East Asian Width]  (MIT)
* [MathJax]  (Apache)
* [highlight.js]  (BSD)


[Go]: http://golang.org/
[blackfriday]: https://github.com/russross/blackfriday
[go-logging]: https://github.com/op/go-logging
[Twitter Bootstrap]: http://getbootstrap.com
[jQuery]: http://jquery.com
[Ace]: http://ace.c9.io
[East Asian Width]: https://github.com/komagata/eastasianwidth
[Mathjax]: http://www.mathjax.org/
[highlight.js]: https://highlightjs.org/
