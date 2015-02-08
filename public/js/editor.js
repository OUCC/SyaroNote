(function(global) {

  var editor,
      modified   = false,
      theme      = 'ace/theme/chrome',
      preview    = true,
      syncScroll = true,
      mathjax    = false,
      timeoutId  = "";

  function init() {
    initAce();
    initUi();
    initTool();
    updateEmojify();
    promptBackup();
  }

  function initUi() {
    $('.alert').hide()
    $('.modal').hide()

    //== modal
    $('#saveButton').on('click', function () {
      if (syaro.gitmode) {
        // restore user signature from local storage
        var name  = localStorage.getItem("name"),
            email = localStorage.getItem("email");
        if (name == undefined)  name = "";
        if (email == undefined) email = "";
        $('#nameInput').val();
        $('#emailInput').val();

        // show save modal
        $('.alert').hide();
        $('#saveModalButton').toggleClass('disabled', false);
        $('#saveModal').modal('show');

      } else {
        simpleSave();
      }
    });

    // button on modal
    $('#saveModalButton').on('click', function() {
      var message = $('#messageInput').val();
      var name    = $('#nameInput').val();
      // var email   = $('#emailInput').val();
      var email   = "";

      // save to local storage
      localStorage.setItem("name", name);
      localStorage.setItem("email", email);

      var callback = function (req) {
        if(req.readyState === 4) {
          $('#saveModalButton').button('reset')
          $('.alert').hide()

          switch (req.status) {
          case 200:
            $('#saveModal').modal('hide');
            indicateSaved();
            modified = false;
            break

          default:
            $('#saveErrorAlert').html('<strong>Error</strong> ' + req.responseText)
            $('#saveErrorAlert').show()
            break
          }
        }
      };

      saveAndPreview(callback, false, message, name, email);

      $('#saveModalButton').button('loading')
    })

    $('#backupModalButton').on('click', function() {
      editor.getSession().getDocument().setValue(syaro.rawBackup);
      $('#backupModal').modal('hide');
    })

    //== option dropdown on navbar
    $('#optionPreview > span').toggleClass('glyphicon-check', true);
    $('#optionSyncScroll > span').toggleClass('glyphicon-check', true);
    $('#optionMathJax > span').toggleClass('glyphicon-unchecked', true);

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

    //== alart
    $(window).on('beforeunload', function () {
      if (modified) {
        return 'Document will not be saved. OK?'
      }
    })
  }

  function initAce() {
    editor = ace.edit('editor')

    editor.setTheme(theme)
    editor.getSession().setMode('ace/mode/markdown')
    editor.getSession().setTabSize(4)
    editor.getSession().setUseSoftTabs(true)
    editor.getSession().setUseWrapMode(true)
    editor.setHighlightActiveLine(true)
    editor.setShowPrintMargin(true)
    editor.setShowInvisibles(true)
    editor.setOption('scrollPastEnd', true)

    editor.getSession().on('change', function(e) {
      modified = true;
      if(timeoutId !== "") { clearTimeout(timeoutId); }

      timeoutId = setTimeout(function() {
        var callback = function (req) {
          if (req.readyState === 4 && req.status === 200) {
            if (!preview) { return; }

            $('#preview').html(req.responseText);
            updateEmojify();

            if (syaro.highlight) {
              $('#preview pre code').each(function(i, block) {
                hljs.highlightBlock(block);
              });
            }

            if (syaro.mathjax && mathjax) {
              // update math in #preview
              MathJax.Hub.Queue(["Typeset", MathJax.Hub, "preview"]);
            }
          }
        }
        saveAndPreview(callback, true);
      }, 2000);
    })

    // sync scroll
    editor.getSession().on('changeScrollTop', scroll)

    // Ctrl-S
    HashHandler = ace.require('ace/keyboard/hash_handler').HashHandler;
    editor.keyBinding.addKeyboardHandler(new HashHandler([{
      bindKey: "Ctrl+S",
      descr:   "Save document",
      exec:    function () {
        if (syaro.gitmode) {
          $('#saveModal').modal('show');
        } else {
          simpleSave();
        }
      },
    }]));
  }

  function initTool() {
    TableFormatter = global['TableFormatter'];
    EmojiAutoComplete = global['EmojiAutoComplete'];

    new TableFormatter(editor);
    new EmojiAutoComplete(editor);
  }

  function promptBackup () {
    if (syaro.rawBackup != undefined) {
      $('#backupModal').modal('show');
    }
  }

  function saveAndPreview(callback, backup, message, name, email) {
    var reqUrl = location.href.split('?')[0];
    if (backup) {
      reqUrl += '?backup=true';
    } else {
      reqUrl += '?message=' + encodeURIComponent(message) +
                '&name='    + encodeURIComponent(name) +
                '&email='   + encodeURIComponent(email);
    }

    var req = new XMLHttpRequest();
    req.open('PUT', reqUrl);

    req.onreadystatechange = function () {
      callback(req);
    };

    req.send(editor.getSession().getValue());
  }

  function indicateSaved() {
    var savedText = "<strong>Saved!</strong>";
    $('#saveButton').html($('#saveButton').html().replace("Save", savedText));

    var count = 9;
    var func = function () {
      if (count % 2 === 1) {
        $('#saveButton').css('color', '#00e613');
      } else {
        $('#saveButton').removeAttr('style');
      }

      if (count !== 0) { setTimeout(func, 333); }
      else {
        $('#saveButton').html($('#saveButton').html().replace(savedText, "Save"));
      }
      count--;
    };
    setTimeout(func, 333)
    editor.focus()
  }

  function simpleSave() {
    var callback = function (req) {
      if(req.readyState === 4) {
        switch (req.status) {
        case 200:
          indicateSaved();
          modified = false;
          break

        default:
          window.alert("**Error** " + req.responseText)
          break
        }
      }
    };
    saveAndPreview(callback, false);
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

  function updateEmojify() {
    emojify.run($('#preview').get(0));
  }

  init()
})((this || 0).self || global);
