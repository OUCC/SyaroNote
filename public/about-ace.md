# About ace
Current version is v1.2.2.

1. Clone https://github.com/ajaxorg/ace
2. `patch -p1 < ace-markdown.patch`
3. `node ./Makefile.dryice.js -m -nc`
4. Copy `ace.js`, `ext-language_tools.js`, `ext-searchbox.js`, `mode-markdown.js`
  and `theme-chrome.js` in `build/src-min-noconflict` to `public/js/ace`
