import * as ace from 'brace'

var HashHandler = ace.acequire('ace/keyboard/hash_handler').HashHandler;
var Range = ace.acequire('ace/range').Range;

import TypeMath from './lib/typemath'

var editor,
    typemath,
    inline,
    mathRange;

export default function mathEditor(edit) {
  editor = edit;
  typemath = new TypeMath($("#mdlMath-field"), $("#mdlMath-latex"),
    $("#mdlMath-candy"), $("#mdlMath-ghost"), $("#mdlMath-select"));

  $('#mdlMath').on('shown.bs.modal', function () {
    $('#mdlMath-latex').focus();
  });
  $('#mdlMath').on('hidden.bs.modal', function () {
    editor.focus();
  });

  $('#mdlMath-insert').on('click', function () {
    $('#mdlMath').modal('hide');
    insertMath();
  });
  $('#mdlMath-discard').on('click', function () {
    $('#mdlMath').modal('hide');
    typemath.setLaTeXCode('');
  });

  $('#dropdown-math').on('click', detectAndShow);

  editor.keyBinding.addKeyboardHandler(new HashHandler([{
    bindKey : 'Ctrl-M',
    descr   : "Math editor",
    exec    : detectAndShow,
  },
  {
    bindKey : 'Ctrl-Shift-M',
    descr   : "Math editor (display math)",
    exec    : showEditor,
  }]));
}

function detectAndShow() {
  detectMath();
  if (mathRange) {
    var code = editor.getSession().getTextRange(mathRange);
    if (inline) {
      code = code.slice(1, code.length-1);
    } else {
      code = code.slice(2, code.length-2);
    }
    typemath.setLaTeXCode(code);
    $('#mdlMath-insert').text('Replace');
  } else {
    inline = true;
    typemath.setLaTeXCode('');
    $('#mdlMath-insert').text('Insert');
  }
  $('#mdlMath').modal('show');
}

function showEditor() {
  inline = false;
  typemath.setLaTeXCode('');
  $('#mdlMath-insert').text('Insert');
  $('#mdlMath').modal('show');
}

function detectMath () {
  var sess = editor.getSession();
  var pos = editor.getCursorPosition();
  var line = sess.getLine(pos.row);

  // inline math $x$
  var c1 = (line.slice(0, pos.column).match(/\$/g) || []).length;
  var c2 = (line.slice(pos.column).match(/\$/g) || []).length;
  if (c1%2 == 1 && c2 > 0) {
    inline = true;
    var startCol = line.slice(0, pos.column).lastIndexOf('$');
    var endCol = pos.column + line.slice(pos.column).indexOf('$') + 1;
    mathRange = new Range(pos.row, startCol, pos.row, endCol);
    return;
  }

  // display math $$x$$
  c1 = (line.slice(0, pos.column).match(/\$\$/g) || []).length;
  c2 = (line.slice(pos.column).match(/\$\$/g) || []).length;
  if (c1 > 0 && c2 > 0) {
    inline = false;
    var startCol = line.slice(0, pos.column).lastIndexOf('$$');
    var endCol = pos.column + line.slice(pos.column).indexOf('$$') + 2;
    mathRange = new Range(pos.row, startCol, pos.row, endCol);
    return;
  }

  // not found
  mathRange = null;
}

function insertMath () {
  var sess = editor.getSession();
  var pos = editor.getCursorPosition();
  var text = typemath.getLaTeXCode();

  text = inline ? '$' + text + '$' : '$$' + text + '$$';

  if (mathRange) {
    sess.replace(mathRange, text);
    editor.moveCursorTo(mathRange.start.row, mathRange.start.column + text.length);
  } else {
    sess.insert(pos, text);
    editor.moveCursorTo(pos.row, pos.column + text.length);
  }
}
