/* global convert */
/* global emojione */
/* global postMessage */ // supress tsc warning

importScripts('/js/convert.min.js');
importScripts('/js/emojione.min.js');

addEventListener('message', (e) => {
  let data = e.data;
  switch (data.cmd) {
  case 'convert':
    console.time('worker.js convert');
    let html = data.data;
    html = convert(html);
    html = emojione.toImage(html);
    postMessage({
      cmd: 'convert',
      data: html,
    });
    console.timeEnd('worker.js convert');
    break;

  case 'diff':
    console.time('worker.js diff');
    postMessage({
      cmd: 'diff',
      data: diff(data.tree, data.newTree),
    });
    console.timeEnd('worker.js diff');
    break;
  }
}, false);
