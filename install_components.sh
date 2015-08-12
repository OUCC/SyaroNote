#!/bin/bash

rm -rf public/lib
mkdir -p public/lib/{js,css,images/emoji}
cp bower_components/jquery/dist/jquery.min.js public/lib/js/jquery.min.js
cp bower_components/jquery/dist/jquery.min.map public/lib/js/jquery.min.map
cp bower_components/ace-builds/src-min-noconflict/ace.js public/lib/js/ace.js
cp bower_components/ace-builds/src-min-noconflict/theme-* public/lib/js/
cp bower_components/ace-builds/src-min-noconflict/mode-markdown.js public/lib/js/mode-markdown.js
cp bower_components/emojify.js/emojify.min.js public/lib/js/emojify.min.js
cp bower_components/emojify.js/images/emoji/* public/lib/images/emoji/
cp eastasianwidth/eastasianwidth.js public/lib/js/eastasianwidth.js
cp bower_components/react/react.min.js public/lib/js/react.min.js
