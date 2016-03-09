/* global syaro */
/* global hljs */
/* global emojione */
import renderPreview from './preview'
import tableFormatter from './tableformatter'
// import emojiAutoComplete from './emojiautocomplete'
import EmojiCompleter from './emojicompleter'
import mathEditor from './matheditor'
import * as api from './api'

const BACKUP_KEY = 'syaro_backup';

var editor
  , math
  , modified = false
  , timeoutId = 0
  , wikiPath = ''
  , fileName = ''
  , savedText = ''
  , initialized = false
  , optionPreview = true
  , optionSyncScroll = true
  , optionBlackTheme = false;

function init() {
  // set wiki path
  wikiPath = get_url_vars()["wpath"];
  var pathlist = wikiPath.split('/');
  fileName = pathlist[pathlist.length-1];

  // update title
  document.title = fileName;

  // update .topbar-header
  $('.topbar-brand').attr('href', wikiPath);
  $('.topbar-brand').text(fileName);

  initUi();
  initAce();

  toastr.options = {
    'positionClass' : 'toast-bottom-right',
  };

  // emojione config
  if (emojione) {
    emojione.unicodeAlt = false;
    emojione.imagePathPNG = '/images/emojione/';
  }

  // load markdown
  api.get(wikiPath)
    .then((arg) => {
      savedText = arg.responseText;

      if (getBackup()) { // backup is available
        $('#mdlBackup').modal({keyboard: false});
      } else { // DONT OVERWRITE BACKUP UNTIL USER SELECTS DISCARD
        editor.getSession().setValue(savedText);
        editor.focus();
        renderPreview(savedText);
        $('#splash').remove();
      }
    })
    .catch((arg) => {
      $('#splash').remove();
      window.alert("**ERROR** failed to load " + wikiPath + "\n" +
        arg.status + " " + arg.statusText + "\n" +
        arg.responseText);
    });
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
  $('.alert').hide();
  $('.modal').hide();

  //
  // topbar
  //
  $('#btnSave').on('click', save);
  $('#btnClose').on('click', () => {
    window.location.href = decodeURIComponent(wikiPath);
  });

  //
  // modal
  //
  $('#mdlSave-save').on('show.bs.modal', () => {
    // restore user signature from local storage
    let name = localStorage.getItem("name")
      , email = localStorage.getItem("email");
    if (name) $('#nameInput').val(name);
    if (email) $('#emailInput').val(email);
  });
  $('#mdlSave-save').on('shown.bs.modal', () => {
    $('#nameInput').focus();
  });
  $('#mdlSave-save').on('click', () => {
    var contents = editor.getSession().getValue();
    var message = $('#messageInput').val();
    var name    = $('#nameInput').val();
    var email   = $('#emailInput').val();

    // save to local storage
    localStorage.setItem("name", name);
    localStorage.setItem("email", email);

    $('#mdlSave').modal('hide');
    toastr.info("", "Saving...");

    api.update(wikiPath, contents, message, name, email)
      .then((arg) => {
        toastr.clear();

        toastr.success("", "Saved");
        setModified(false);
        removeBackup();
        savedText = contents;
      })
      .catch((arg) => {
        toastr.clear();
        toastr.error(arg.status + " " + arg.statusText + "\n" +
          arg.responseText, "ERROR!");
      });
  });
  $('#mdlBackup-restore').on('click', () => {
    $('#mdlBackup').modal('hide');
    setModified(true);

    let text = getBackup();
    editor.getSession().setValue(text);
    editor.focus();
    renderPreview(text);
    $('#splash').remove();
  });
  $('#mdlBackup-discard').on('click', () => {
    removeBackup();
    $('#mdlBackup').modal('hide');
    setModified(false);

    editor.getSession().setValue(savedText);
    editor.focus();
    renderPreview(savedText);
    $('#splash').remove();
  });

  //
  // option dropdown on topbar
  //
  $('#optionPreview > span').toggleClass('glyphicon-check', true);
  $('#optionSyncScroll > span').toggleClass('glyphicon-check', true);
  $('#optionBlackTheme > span').toggleClass('glyphicon-unchecked', true);

  $('#optionPreview').on('click', function () {
    optionPreview = !optionPreview;
    $('#optionPreview > span').toggleClass('glyphicon-check');
    $('#optionPreview > span').toggleClass('glyphicon-unchecked');
    $('#optionMathJax').parent('li').toggleClass('disabled');

    if (optionPreview) {
      let markdown = editor.getSession().getValue();
      renderPreview(markdown);
    }
    return false;
  });

  $('#optionSyncScroll').on('click', function () {
    optionSyncScroll = !optionSyncScroll;
    $('#optionSyncScroll > span').toggleClass('glyphicon-check');
    $('#optionSyncScroll > span').toggleClass('glyphicon-unchecked');
    return false;
  });

  $('#optionBlackTheme').on('click', () => {
    optionBlackTheme = !optionBlackTheme;
    $('#optionBlackTheme > span').toggleClass('glyphicon-check');
    $('#optionBlackTheme > span').toggleClass('glyphicon-unchecked');

    if (optionBlackTheme) {
      $(document.body).toggleClass('theme-black', true);
      editor.setTheme('ace/theme/monokai');
    } else {
      $(document.body).toggleClass('theme-black', false);
      editor.setTheme('ace/theme/chrome');
    }
    return false;
  });

  //
  // alert
  //
  $(window).on('beforeunload', function () {
    if (modified) {
      return 'Document will not be saved. OK?';
    }
  });
}

function initAce() {
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
    scrollPastEnd: true,
    cursorStyle: 'smooth', // "ace"|"slim"|"smooth"|"wide"
    enableBasicAutocompletion: true,
    enableLiveAutocompletion: true,
  });

  // disable message
  // Automatically scrolling cursor into view after selection change this will be disabled in the next version
  editor.$blockScrolling = Infinity;

  //
  // Keyboard shortcut
  //
  // Ctrl-S: save
  var HashHandler = ace.require('ace/keyboard/hash_handler').HashHandler;
  editor.keyBinding.addKeyboardHandler(new HashHandler([{
    bindKey: "Ctrl-S",
    descr:   "Save document",
    exec: save,
  }]));

  // register ace HashHandlers
  tableFormatter(editor);
  // emojiAutoComplete(editor);
  mathEditor(editor);
  EmojiCompleter();

  //
  // Event
  //
  editor.getSession().on('change', (e) => {
    if (!initialized) {
      initialized = true;
      return;
    }
    setModified(true);

    if (timeoutId) { clearTimeout(timeoutId); }

    timeoutId = setTimeout(() => {
      backup();
      timeoutId = 0;

      if (!optionPreview) { return; }

      let markdown = editor.getSession().getValue();
      renderPreview(markdown);
    }, 600);
  });

  // sync scroll
  editor.getSession().on('changeScrollTop', syncScroll);
}

function setModified(b) {
  if (b) {
    modified = true;
    document.title = '* ' + fileName; // update title
    $('#btnSave').addClass('modified');
  } else {
    modified = false;
    document.title = fileName;
    $('#btnSave').removeClass('modified');
  }
}

function save() {
  if (syaro.gitmode) {
    $('#mdlSave').modal('show');
    return;
  }

  toastr.info("Saving...");

  let contents = editor.getSession().getValue();
  api.update(wikiPath, contents, null, null, null)
    .then((arg) => {
      toastr.clear();
      toastr.success("", "Saved");
      modified = false;
      document.title = fileName;
      $('#btnSave').removeClass('modified');
      removeBackup();
      savedText = contents;
    })
    .catch((arg) => {
      toastr.clear();
      toastr.error(arg.status + " " + arg.statusText + "\n" +
        arg.responseText, "ERROR!");
    });
}

function backup() {
  let key = BACKUP_KEY + '_' + wikiPath
    , contents = editor.getSession().getValue();
  localStorage.setItem(key, contents);
}

function getBackup() {
  let key = BACKUP_KEY + '_' + wikiPath;
  return localStorage.getItem(key);
}

function removeBackup() {
  let key = BACKUP_KEY + '_' + wikiPath;
  localStorage.removeItem(key);
}

function syncScroll() {
  if (!optionSyncScroll) { return; }

  var $preview = $('#preview');

  var previewHeight  = $preview[0].scrollHeight,
      previewVisible = $preview.height(),
      // previewTop     = $preview[0].scrollTop,
      editorHeight   = editor.getSession().getLength(),
      editorVisible  = editor.getLastVisibleRow() - editor.getFirstVisibleRow(),
      editorTop      = editor.getFirstVisibleRow();

  // editorTop / (editorHeight - editorVisible)
  //   = previewTop / (previewHeight - previewVisible)
  var top = editorTop * (previewHeight - previewVisible) / (editorHeight - editorVisible);

  // $preview.scrollTop(top);
  $preview.animate({ scrollTop: top }, 10, 'swing');
}

init();
