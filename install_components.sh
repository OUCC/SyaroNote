#!/bin/bash

rm -rf public/lib
mkdir -p public/lib/{jquery,bootstrap,ace,marked}
cp bower_components/jquery/dist/jquery.min.js public/lib/jquery/jquery.min.js
cp bower_components/bootstrap/dist/css/bootstrap.min.css public/lib/bootstrap/bootstrap.min.css
cp bower_components/bootstrap/dist/fonts/* public/lib/bootstrap/
cp bower_components/bootstrap/dist/js/bootstrap.min.js public/lib/bootstrap/bootstrap.min.js
cp bower_components/ace-builds/src-min-noconflict/ace.js public/lib/ace/ace.js
cp bower_components/ace-builds/src-min-noconflict/theme-* public/lib/ace/
cp bower_components/ace-builds/src-min-noconflict/mode-markdown.js public/lib/ace/mode-markdown.js
cp bower_components/marked/lib/marked.js public/lib/marked/marked.js
