#!/bin/bash

if [ -d build ]; then
    rm -rf build
fi
mkdir build

# syaro
go build -ldflags "-X main.version=$(git describe)" -o build/syaro

# copy public
cd $GOPATH/src/github.com/OUCC/syaro/
cp -a public build/public

# copy template
cp -a template build/template

# bower
bower update
mkdir -p build/public/{js,css,images/sprites}
cp bower_components/jquery/dist/jquery.min.js build/public/js/jquery.min.js
cp bower_components/jquery/dist/jquery.min.map build/public/js/jquery.min.map
cp bower_components/emojify.js/dist/js/emojify.min.js build/public/js/emojify.min.js
cp bower_components/emojify.js/dist/css/sprites/emojify.min.css build/public/css/emojify.sprites.min.css
cp bower_components/emojify.js/dist/css/sprites/emojify-emoticons.min.css build/public/css/emojify-emoticons.sprites.min.css
cp bower_components/emojify.js/dist/css/basic/emojify.min.css build/public/css/emojify.basic.min.css
cp -a bower_components/emojify.js/dist/images/sprites build/public/images/sprites
cp -a bower_components/emojify.js/dist/images/basic build/public/images/emoji
cp bower_components/toastr/toastr.min.js build/public/js/toastr.min.js
cp bower_components/toastr/toastr.min.css build/public/css/toastr.min.css

# editor.js
# git submodule update --init # eastasianwidth
cd editorjs
if [ -d build ]; then
    rm -rf build
fi
mkdir build
npm install
npm run deploy

# gopherjs
gopherjs build -m -o ../build/public/js/convert.min.js convert.go

exit
# page.js
cd ../pagejs
if [ -d build ]; then
    rm -rf build
fi
mkdir build
npm install
npm run deploy
