#!/bin/bash

rm -rf public/lib
mkdir -p public/lib/{js,css}
cp bower_components/jquery/dist/jquery.min.js public/lib/js/jquery.min.js
cp bower_components/jquery/dist/jquery.min.map public/lib/js/jquery.min.map
cp bower_components/bootstrap/dist/css/bootstrap.min.css public/lib/css/bootstrap.min.css
cp -a bower_components/bootstrap/dist/fonts public/lib/
cp bower_components/bootstrap/dist/js/bootstrap.min.js public/lib/js/bootstrap.min.js
cp bower_components/ace-builds/src-min-noconflict/ace.js public/lib/js/ace.js
cp bower_components/ace-builds/src-min-noconflict/theme-* public/lib/js/
cp bower_components/ace-builds/src-min-noconflict/mode-markdown.js public/lib/js/mode-markdown.js
cp bower_components/marked/lib/marked.js public/lib/js/marked.js
