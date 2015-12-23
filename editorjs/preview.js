/* global hljs */
/* global emojify */
import {diff, patch} from 'virtual-dom'
import virtualize from 'vdom-virtualize'

var preview = $('#preview div').get(0)
  , buffer = $('#preview-buffer div').get(0)
  , running = false;

export function render(html) {
  if (running) { return; }
  running = true;

  console.debug('rendering preview...');

  buffer.innerHTML = html;

  if (emojify) {
    emojify.run(buffer); // sync
  }

  if (hljs) {
    $('#preview-buffer pre code').each(function(i, block) {
      hljs.highlightBlock(block); // sync
    });
  }

  // http://mathjax.readthedocs.org/en/latest/typeset.html
  if (MathJax) {
    MathJax.Hub.Queue(['Typeset', MathJax.Hub, 'preview-buffer'],
      applyChanges);
    // MathJax.Hub.Queue(diffAndPatch);
  } else {
    applyChanges();
  }
}

function applyChanges() {
  var tree = virtualize(preview);
  var newTree = virtualize(buffer);
  var patches = diff(tree, newTree);
  patch(preview, patches);
  running = false;

  console.debug('applyChanges end');
}
