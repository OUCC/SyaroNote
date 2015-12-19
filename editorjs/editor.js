/* global ace */
import tableFormatter from './tableformatter'
import emojiAutoComplete from './emojiautocomplete'
import mathEditor from './matheditor'
import * as api from './api'

const BACKUP_KEY = 'syaro_backup';

var editor,
    math,
    initialized = false,
    modified = false,
    timeoutId = '',
    wikiPath = '',
    fileName = '',
    syncScroll = true;

function init() {
  // set wiki path
  wikiPath = get_url_vars()["wpath"];
  fileName = wikiPath.split('/');
  fileName = fileName[fileName.length-1];

  // update title
  document.title = fileName;

  // update .navbar-header
  $('.navbar-brand').attr('href', wikiPath);
  $('.navbar-brand').text(fileName);

  initUi();
  initEmojify();
  initAce();

  toastr.options = {
    'positionClass' : 'toast-bottom-right',
  };

  // register ace HashHandlers
  tableFormatter(editor);
  emojiAutoComplete(editor);
  mathEditor(editor);

  // load markdown
  api.get(wikiPath, function (contents, err) {
    if (contents) { editor.getSession().setValue(contents); }
    else { window.alert("**ERROR** failed to load " + wikiPath + "\n" + err); }
    editor.focus();
  });

  // TODO backup
  if (getBackup()) {
    $('#mdlBackup').modal('show');
  }
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
        document.title = fileName;
        $('#btnSave').removeClass('modified');
        backup(true);
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
    editor.getSession().getDocument().setValue(getBackup());// FIXME
    $('#mdlBackup').modal('hide');
  });
  $('#mdlBackup-discard').on('click', function() {
    backup(true); // remove backup
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
  ace.require("ace/ext/language_tools");
  editor = ace.edit('editor');

  editor.setTheme('ace/theme/chrome');
  editor.getSession().setMode('ace/mode/markdown');
  editor.getSession().setTabSize(2);
  editor.getSession().setUseSoftTabs(true);
  editor.getSession().setUseWrapMode(true);
  editor.setHighlightActiveLine(true);
  editor.setShowInvisibles(true);
  editor.session.setNewLineMode('unix');
  editor.setOptions({
    'scrollPastEnd': true,
    cursorStyle: 'smooth', // "ace"|"slim"|"smooth"|"wide"
    enableBasicAutocompletion: true,
    enableLiveAutocompletion: true,
  });

  // disable message
  // Automatically scrolling cursor into view after selection change this will be disabled in the next version
  editor.$blockScrolling = Infinity;

  editor.getSession().on('change', (e) => {
    modified = true;

    // update title
    document.title = '* '+fileName;

    $('#btnSave').addClass('modified');

    if(timeoutId !== "") { clearTimeout(timeoutId); }

    timeoutId = setTimeout(() => {
      if (initialized) {
        backup(false);
      }
      renderPreview();
    }, 600);
  })

  // sync scroll
  editor.getSession().on('changeScrollTop', scroll)

  // Ctrl-S: save
  var HashHandler = ace.require('ace/keyboard/hash_handler').HashHandler;
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
  var html = convert(editor.getSession().getValue());
  $('#preview').html(html);

  emojify.run($('#preview').get(0));

  if (syaro.highlight && hljs) {
    $('#preview pre code').each(function(i, block) {
      hljs.highlightBlock(block);
    });
  }

  // http://mathjax.readthedocs.org/en/latest/typeset.html
  if (syaro.mathjax && MathJax) {
    // update math in #preview
    MathJax.Hub.Queue(["Typeset", MathJax.Hub, "preview"]);
  }

  if (!initialized) {
    $('#splash').remove();
    initialized = true;
    modified = false;
    document.title = fileName;
    $('#btnSave').removeClass('modified');
  }
}

function simpleSave() {
  var callback = function (err) {
    toastr.clear();
    if (!err) {
      toastr.success("", "Saved");
      modified = false;
      document.title = fileName;
      $('#btnSave').removeClass('modified');
      backup(true);
    } else {
      toastr.error(err, "Error!");
    }
  };
  api.update(wikiPath, editor.getSession().getValue(), callback);
  toastr.info("Saving...");
}

function backup(remove) {
  let key = BACKUP_KEY+'_'+wikiPath;
  if (remove) {
    localStorage.removeItem(key);
  } else {
    localStorage.setItem(key, editor.getSession().getValue());
  }
}

function getBackup() {
  let key = BACKUP_KEY+'_'+wikiPath;
  return localStorage.getItem(key);
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
