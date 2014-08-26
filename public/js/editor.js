$(function() {

  var editor
  var modified = false
  var theme = 'ace/theme/chrome'
  var preview = true
  var mathjax = false

  function init() {
    initUi()
    initAce()
    initMarked()
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
      modified = true
      updatePreview()
    })

    // sync scroll
    editor.getSession().on('changeScrollTop', syncScroll)
  }

  function initMarked() {
    marked.setOptions({
      gfm: true,
      tables: true,
      pedantic: false,
      sanitize: false,
      smartLists: true,
      smartypants: false,
      langPrefix: 'lang-',
      highlight: syaro.highlight ? function(code) { return hljs.highlightAuto(code).value; }
                                 : undefined
    })
  }

  function updatePreview() {
    if (!preview) { return }

    var md = editor.getSession().getValue()
    var mdhtml = marked(md)

    $('#preview').html(mdhtml)

    if (syaro.mathjax && mathjax) {
      // update math
      MathJax.Hub.Queue(["Typeset", MathJax.Hub])
    }
  }

  function syncScroll() {
    var $prev = $('#preview')

    var editorHeight = editor.getSession().getLength()
    var previewHeight = $prev[0].scrollHeight

    // Find how far along the editor is (0 means it is scrolled to the top, 1
    // means it is at the bottom).
    var scrollFactor = editor.getFirstVisibleRow() / editorHeight

    // Set the scroll position of the preview pane to match.  jQuery will
    // gracefully handle out-of-bounds values.
    $prev.scrollTop(scrollFactor * previewHeight)
  }

  init()
})
