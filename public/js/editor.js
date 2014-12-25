(function(global) {

  var editor
  var modified = false
  var theme = 'ace/theme/chrome'
  var preview = true
  var mathjax = false
  var timeoutId = ""

  function init() {
    initUi()
    initAce()
    initTableFormatter();
    updatePreview()
  }

  function initUi() {
    $('.alert').hide()
    $('.modal').hide()

    $('#saveModal').on('show.bs.modal', function() {
      $('.alert').hide()
      $('#saveModalButton').toggleClass('disabled', false)
    })

    // option dropdown on navbar
    $('#optionPreview > span').toggleClass('glyphicon-check', true)
    $('#optionMathJax > span').toggleClass('glyphicon-unchecked', true);

    $('#optionPreview').on('click', function() {
      preview = !preview
      $('#optionPreview > span').toggleClass('glyphicon-check')
      $('#optionPreview > span').toggleClass('glyphicon-unchecked')
      $('#optionMathJax').parent('li').toggleClass('disabled')
      return false
    })

    $('#optionMathJax').on('click', function() {
      mathjax = !mathjax
      $('#optionMathJax > span').toggleClass('glyphicon-check')
      $('#optionMathJax > span').toggleClass('glyphicon-unchecked')
      return false
    })

    // button on navbar
    $('a.close-button').on('click', function() {
      if (modified) {
        $('#closeModal').modal()
      } else {
        // back to page view
        location.href = location.href.split('?')[0]
      }
      return false
    })

    // button on modal
    $('#saveModalButton').on('click', function() {
      var req = new XMLHttpRequest()
      req.open('POST', location.href)

      req.onreadystatechange = function() {
        if(req.readyState === 4) {
          $('#saveModalButton').button('reset')
          $('.alert').hide()

          switch (req.status) {
          case 200:
            $('#savedAlert').show()
            modified = false
            $('#saveModalButton').toggleClass('disabled', true) // FIXME don't work!
            break

          default:
            $('#saveErrorAlert').html('<strong>Error</strong> ' + req.responseText)
            $('#saveErrorAlert').show()
            break
          }
        }
      }

      req.send(editor.getSession().getValue())
      $('#saveModalButton').button('loading')
    })

    $('#closeModalButton').on('click', function() {
        // back to page view
        location.href = location.href.split('?')[0]
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
        updatePreview();
      }, 5000);
    })

    // sync scroll
    editor.getSession().on('changeScrollTop', syncScroll)
  }

  function initTableFormatter() {
    // http://stackoverflow.com/questions/14042926/keydown-event-not-fired-on-ace-editor
    HashHandler = ace.require('ace/keyboard/hash_handler').HashHandler;
    TableFormatter = global['TableFormatter'];

    editor.keyBinding.addKeyboardHandler(new TableFormatter());
  }

  function updatePreview() {
    if (!preview) { return; }

    var url = document.createElement('a');
    url.href = location.href;

    var reqUrl = url.protocol + '//' + url.host +
        syaro.urlPrefix === '' ? '/' + syaro.urlPrefix : '' +
        '/preview?path=' + encodeURIComponent(syaro.wikiPath).replace(/%2F/g, '/');

    var req = new XMLHttpRequest();
    req.open('POST', reqUrl);

    req.onreadystatechange = function() {
      if(req.readyState === 4 && req.status === 200) {
        $('#preview').html(req.responseText);

        if (syaro.highlight) {
          $('#preview pre code').each(function(i, block) {
            hljs.highlightBlock(block);
          });
        }

        if (syaro.mathjax && mathjax) {
          // update math in #preview
          MathJax.Hub.Queue(["Typeset", MathJax.Hub, "preview"]);
        }

        syncScroll();
      }
    }

    req.send(editor.getSession().getValue());
  }

  function syncScroll() {
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

  init()
})((this || 0).self || global);
