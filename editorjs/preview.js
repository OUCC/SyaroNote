/* global hljs */
/* global emojione */
import {diff, patch} from 'virtual-dom'
import virtualize from 'vdom-virtualize'

var worker = new Worker('/js/worker.js')
  , preview = $('#preview div').get(0)
  , buffer = $('#preview-buffer div').get(0)
  , oldTree
  , running = false;

oldTree = virtualize(preview);

worker.addEventListener('message', (e) => {
  let data = e.data;
  switch (data.cmd) {
  case 'convert':
    updateBuffer(data.data);
    break;

  case 'diff':
    // applyChanges(data.data);
    break;
  }
});

export default function renderPreview(markdown) {
  if (running) { return; }
  running = true;

  console.time('renderPreview');

  worker.postMessage({ cmd: 'convert', data: markdown });
}

function updateBuffer(html) {
  console.time('updateBuffer innerHTML');
  buffer.innerHTML = html;
  console.timeEnd('updateBuffer innerHTML');

  console.time('updateBuffer hljs');
  if (hljs) {
    $('#preview-buffer pre code').each(function(i, block) {
      hljs.highlightBlock(block); // sync
    });
  }
  console.timeEnd('updateBuffer hljs');

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
  console.time('applyChanges');

  console.time('applyChanges virtualize');
  // var tree = virtualize(preview);
  var newTree = virtualize(buffer);
  console.timeEnd('applyChanges virtualize');

  // worker.postMessage({ cmd: 'diff', tree: tree, newTree: newTree });
  console.time('applyChanges diff');
  var patches = diff(oldTree, newTree);
  console.timeEnd('applyChanges diff');

  console.time('applyChanges patch');
  patch(preview, patches);
  console.timeEnd('applyChanges patch');

  running = false;
  oldTree = newTree;

  console.timeEnd('applyChanges');
  console.timeEnd('renderPreview');
}
