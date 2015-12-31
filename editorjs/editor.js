/* global syaro */
/* global hljs */
/* global convert */
/* global emojione */
import * as preview from './preview'
import tableFormatter from './tableformatter'
// import emojiAutoComplete from './emojiautocomplete'
import EmojiCompleter from './emojicompleter'
import mathEditor from './matheditor'
import * as api from './api'

const BACKUP_KEY = 'syaro_backup';

var editor,
    math,
    modified = false,
    timeoutId = 0,
    wikiPath = '',
    fileName = '',
    savedText = '',
    initialized = false,
    optionPreview = true,
    optionSyncScroll = true;

function init() {
  // set wiki path
  wikiPath = get_url_vars()["wpath"];
  var pathlist = wikiPath.split('/');
  fileName = pathlist[pathlist.length-1];

  // update title
  document.title = fileName;

  // update .navbar-header
  $('.navbar-brand').attr('href', wikiPath);
  $('.navbar-brand').text(fileName);

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
      $('#splash').remove();

      if (getBackup()) { // backup is available
        $('#mdlBackup').modal({keyboard: false});
      } else { // DONT OVERWRITE BACKUP UNTIL USER SELECTS DISCARD
        editor.getSession().setValue(savedText);
        editor.focus();
        initialized = true;
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
  // navbar
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
  });
  $('#mdlBackup-restore').on('click', () => {
    $('#mdlBackup').modal('hide');

    initialized = true;
    editor.getSession().setValue(getBackup());
    editor.focus();
  });
  $('#mdlBackup-discard').on('click', () => {
    removeBackup();
    $('#mdlBackup').modal('hide');

    editor.getSession().setValue(savedText);
    editor.focus();
    initialized = true;
  });

  //
  // option dropdown on navbar
  //
  $('#optionPreview > span').toggleClass('glyphicon-check', true);
  $('#optionSyncScroll > span').toggleClass('glyphicon-check', true);

  $('#optionPreview').on('click', function () {
    optionPreview = !optionPreview;
    $('#optionPreview > span').toggleClass('glyphicon-check');
    $('#optionPreview > span').toggleClass('glyphicon-unchecked');
    $('#optionMathJax').parent('li').toggleClass('disabled');

    if (optionPreview) {
      var html = convert(editor.getSession().getValue());
      preview.render(html);
    }
    return false;
  });

  $('#optionSyncScroll').on('click', function () {
    optionSyncScroll = !optionSyncScroll;
    $('#optionSyncScroll > span').toggleClass('glyphicon-check');
    $('#optionSyncScroll > span').toggleClass('glyphicon-unchecked');
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
    if (!initialized) { return; }
    modified = true;
    document.title = '* ' + fileName; // update title
    $('#btnSave').addClass('modified');

    if (timeoutId) { clearTimeout(timeoutId); }

    timeoutId = setTimeout(() => {
      backup();
      timeoutId = 0;

      if (!optionPreview) { return; }

      var html = convert(editor.getSession().getValue());
      preview.render(html);
    }, 600);
  });

  // sync scroll
  editor.getSession().on('changeScrollTop', syncScroll);
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
