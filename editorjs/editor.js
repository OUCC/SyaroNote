import * as ace from 'brace'
import 'brace/mode/markdown'
import 'brace/theme/chrome'
import 'brace/ext/searchbox'
var HashHandler = ace.acequire("ace/keyboard/hash_handler").HashHandler;

import tableFormatter from './tableformatter'
import emojiAutoComplete from './emojiautocomplete'

import h from 'virtual-dom/h'
import diff from 'virtual-dom/diff'
import patch from 'virtual-dom/patch'
import createElement from 'virtual-dom/create-element'
import virtualize from 'vdom-virtualize'

import * as api from './api'

var editor,
    initialized = false,
    modified = false,
    timeoutId = '',
    wikiPath = '',
    syncScroll = true;

function init() {
  // set wiki path
  wikiPath = get_url_vars()["wpath"];

  // update title
  document.title = "Editing " + setting.wikiPath;

  initUi();
  initEmojify();
  initAce();

  toastr.options = {
    'positionClass' : 'toast-bottom-right',
  };

  // register ace HashHandlers
  tableFormatter(editor);
  emojiAutoComplete(editor);

  // load markdown
  api.get(wikiPath, function (contents, err) {
    if (contents) { editor.getSession().setValue(contents); }
    else { window.alert("**ERROR** failed to load " + wikiPath + "\n" + err); }
    modified = false;
    editor.focus();
  });

  // TODO backup
}

function get_url_vars() {
  var vars = new Object, params;
  var temp_params = window.location.search.substring(1).split('&');
  for(var i = 0; i <temp_params.length; i++) {
    params = temp_params[i].split('=');
    vars[params[0]] = decodeURIComponent(params[1]);
  }
  return vars;
}

function initUi() {
  $('.alert').hide()
  $('.modal').hide()

  //
  // navbar
  //
  $('#btnSave').on('click', function () {
    if (syaro.gitmode) {
      // restore user signature from local storage
      var name  = localStorage.getItem("name"),
          email = localStorage.getItem("email");
      if (name) $('#nameInput').val(name);
      if (email) $('#emailInput').val(email);

      // $('.alert').hide();
      // $('#saveModalButton').toggleClass('disabled', false);
      // show save modal
      $('#mdlSave').modal('show');

    } else {
      simpleSave();
    }
  });

  $('#btnClose').on('click', function () {
    window.location.href = decodeURIComponent(wikiPath);
  });

  //
  // modal
  //
  $('#mdlSave-save').on('click', function() {
    var message = $('#messageInput').val();
    var name    = $('#nameInput').val();
    var email   = $('#emailInput').val();

    // save to local storage
    localStorage.setItem("name", name);
    localStorage.setItem("email", email);

    var callback = function (err) {
      // $('#mdlSave-save').button('reset');
      // $('.alert').hide();
      toastr.clear();

      if (!err) {
        // $('#saveModal').modal('hide');
        toastr.success("", "Saved");
        modified = false;
      } else {
        // $('#mdlSave-alart').html('<strong>Error</strong> ' + req.responseText);
        // $('#mdlSave-alart').show();
        toastr.error(err, "Error!");
      }
    };
    api.update(wikiPath, editor.getSession().getValue(), callback,
      message, name, email);

    // $('#saveModalButton').button('loading');
    $('#mdlSave').modal('hide');
    toastr.info("", "Saving...");
  });
  $('#mdlBackup-restore').on('click', function() {
    editor.getSession().getDocument().setValue();// FIXME
    $('#mdlBackup').modal('hide');
  });
  $('#mdlBackup-discard').on('click', function() {
    // TODO
    $('#mdlBackup').modal('hide');
  });

  //
  // option dropdown on navbar
  //
  $('#optionPreview > span').toggleClass('glyphicon-check', true);
  $('#optionSyncScroll > span').toggleClass('glyphicon-check', true);
  // $('#optionMathJax > span').toggleClass('glyphicon-unchecked', true);

  $('#optionPreview').on('click', function() {
    preview = !preview
    $('#optionPreview > span').toggleClass('glyphicon-check')
    $('#optionPreview > span').toggleClass('glyphicon-unchecked')
    $('#optionMathJax').parent('li').toggleClass('disabled')
    return false
  })

  $('#optionSyncScroll').on('click', function() {
    syncScroll = !syncScroll;
    $('#optionSyncScroll > span').toggleClass('glyphicon-check');
    $('#optionSyncScroll > span').toggleClass('glyphicon-unchecked');
    return false
  })

  $('#optionMathJax').on('click', function() {
    mathjax = !mathjax
    $('#optionMathJax > span').toggleClass('glyphicon-check')
    $('#optionMathJax > span').toggleClass('glyphicon-unchecked')
    return false
  })

  //
  // alert
  //
  $(window).on('beforeunload', function () {
    if (modified) {
      return 'Document will not be saved. OK?'
    }
  })
}

function initAce() {
  editor = ace.edit('editor')

  editor.setTheme('ace/theme/chrome')
  editor.getSession().setMode('ace/mode/markdown')
  editor.getSession().setTabSize(4)
  editor.getSession().setUseSoftTabs(true)
  editor.getSession().setUseWrapMode(true)
  editor.setHighlightActiveLine(true)
  editor.setShowPrintMargin(true)
  editor.setShowInvisibles(true)
  editor.setOption('scrollPastEnd', true)

  editor.getSession().on('change', (e) => {
    modified = true;

    if(timeoutId !== "") { clearTimeout(timeoutId); }

    timeoutId = setTimeout(() => {
      renderPreview();
    }, 600);
  })

  // sync scroll
  editor.getSession().on('changeScrollTop', scroll)

  // Ctrl-S: save
  // HashHandler = ace.require('ace/keyboard/hash_handler').HashHandler;
  editor.keyBinding.addKeyboardHandler(new HashHandler([{
    bindKey: "Ctrl-S",
    descr:   "Save document",
    exec:    function () {
      if (syaro.gitmode) {
        $('#mdlSave').modal('show');
      } else {
        simpleSave();
      }
    },
  }]));
}

function initEmojify() {
  emojify.setConfig({
      mode: 'sprites',
      ignore_emoticons: true,
  });
}

function renderPreview() {
  console.debug('rendering preview...');
  $('#preview').html(convert(editor.getSession().getValue()));

  emojify.run($('#preview').get(0));

  if (syaro.highlight && hljs) {
    $('#preview pre code').each(function(i, block) {
      hljs.highlightBlock(block);
    });
  }

  if (syaro.mathjax && MathJax) {
    // update math in #preview
    MathJax.Hub.Queue(["Typeset", MathJax.Hub, "preview"]);
  }

  if (!initialized) {
    $('#splash').remove();
    initialized = true;
    modified = false;
  }
}

function simpleSave() {
  var callback = function (err) {
    toastr.clear();
    if (!err) {
      toastr.success("", "Saved");
      modified = false;
      document.title = fileName;
    } else {
      toastr.error(err, "Error!");
    }
  };
  api.update(wikiPath, editor.getSession().getValue(), callback);
  toastr.info("Saving...");
}

function scroll() {
  if (!syncScroll) { return; }

  var $preview = $('#preview');

  var previewHeight  = $preview[0].scrollHeight,
      previewVisible = $preview.height(),
      previewTop     = $preview[0].scrollTop,
      editorHeight   = editor.getSession().getLength(),
      editorVisible  = editor.getLastVisibleRow() - editor.getFirstVisibleRow(),
      editorTop      = editor.getFirstVisibleRow();

  // editorTop / (editorHeight - editorVisible)
  //   = previewTop / (previewHeight - previewVisible)
  var top = editorTop * (previewHeight - previewVisible) / (editorHeight - editorVisible);

  $preview.scrollTop(top);
}

init();
