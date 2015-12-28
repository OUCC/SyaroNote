/* global emojione */

let langTools = ace.require("ace/ext/language_tools");

// {
//   "grinning": {
//     "unicode": "1f600",
//     "shortname": ":grinning:",
//     "aliases": "",
//     "keywords": "grinning face happy joy smile grin smiling smiley person"
//   },
// }
let emojiStrategy = {};

export default function EmojiCompleter() {
  $.getJSON('/js/emoji_strategy.json', data => {
    emojiStrategy = data;
  });

  let emojiCompleter = {
    getCompletions: (editor, session, pos, prefix, callback) => {
      if (!emojiStrategy) {
        callback(null, []);
        return;
      }

      let line = session.getLine(pos.row).substr(0, pos.column);
      // let m = line.match(/\B:([\w+\-]+)$/);
      if (!line.match(/\B:[\w+\-]+$/)) {
        callback(null, []);
        return;
      }
      let candidates = [];

      $.each(emojiStrategy, (k, v) => {
        if (k.indexOf(prefix) > -1) {
          candidates.push({
            value: k + ': ',
            caption: k,
            meta: 'emoji',
            type: 'emoji',
            score: 3,
            emojiname: k, // not used by ace
          });
        } else if (v.aliases && v.aliases.indexOf(prefix) > -1) {
          let alias = v.aliases.split(' ').filter(v => v.indexOf(prefix) > -1)[0];
          candidates.push({
            // value: k + ': ',
            snippet: k + ': ', // use snippet, otherwise candidate not shown
            caption: alias,
            meta: 'emoji',
            type: 'emoji',
            score: 2,
            emojiname: k,
          });
        } else if (v.keywords && v.keywords.indexOf(prefix) > -1) {
          let keyword = v.keywords.split(' ').filter(v => v.indexOf(prefix) > -1)[0];
          candidates.push({
            snippet: k + ': ',
            caption: keyword,
            meta: 'emoji',
            type: 'emoji',
            score: 1,
            emojiname: k,
          });
        }
      });
      callback(null, candidates);
    },
    getDocTooltip: item => {
      if (item.meta === "emoji" && !item.docHTML) {
        let data = emojiStrategy[item.emojiname];
        item.docHTML = `${emojione.toImage(data.shortname)}<br><b>${data.shortname}</b>`;
      }
    },
  };
  langTools.addCompleter(emojiCompleter);
}
