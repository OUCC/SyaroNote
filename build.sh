#!/bin/bash

if [ -d build ]; then
    rm -rf build
fi
mkdir build

# syaro
go build -ldflags "-X main.version $(git describe)" -o build/syaro

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
cp bower_components/emojify.js/dist/css/sprites/emojify.min.css build/public/css/emojify.min.css
# cp bower_components/emojify.js/dist/images/sprites/emojify.png build/public/images/sprites/emojify.png
convert -geometry 50% bower_components/emojify.js/dist/images/sprites/emojify.png build/public/images/sprites/emojify.png
cp bower_components/emojify.js/dist/images/basic/* build/public/images/

# editor.js
# git submodule update --init # eastasianwidth
cd editorjs
if [ -d build ]; then
    rm -rf build
fi
mkdir build
npm install
npm run deploy

exit
# page.js
cd ../pagejs
if [ -d build ]; then
    rm -rf build
fi
mkdir build
npm install
npm run deploy
