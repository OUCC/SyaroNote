/* global emojify */
/* global syaro */
$(function(){
"use strict";

  var historyLoaded = false,
      reMarkdown = /\.(md|mkd|mkdn|mdown|markdown)$/,
      commErrorText = "Communication error";

  function init() {
    initModal();
    initFileList();
    initFileListMenu();
    initEmojify();
    initHistory();

    if ($('.syaro-right nav')) {
      $(window).scroll(scrollNav);
    }
  }

  function initModal() {
    $('.alert').hide();
    $('.modal').modal('hide');

    $('#createModal').on('show.bs.modal', function (ev) {
      $('#createModal .alert').hide();
      $('#createModal input').val('');
    });

    $('#createModal').on('shown.bs.modal', function (ev) {
      $('#createModal input').focus();
    });

    $('#createModal input').on('keypress', function (ev) {
      if (ev.keyCode === 13) // enter
        $('#createModalButton').trigger('click');
    })

    $('#createModalButton').on('click', function (ev) {
      var filename = $('#createModal input').val();
      if (filename.length === 0) { return; }

      $('#createModalButton').button('loading')
      if (!filename.match(reMarkdown)) {
        filename = filename + '.md';
      }
      if (filename[0] !== '/') {
        filename = syaro.wikiPath + '/' + filename;
      }

      Promise.resolve({
        url: '/api/new?wpath=' + encodeURIComponent(filename),
        method: 'GET',
      }).then(xhr)
      .then(function (arg) {
        // redirect to editor
        location.href = '/edit?wpath=' + encodeURIComponent(filename);
      })
      .catch(function (arg) {
        $('#createModalButton').button('reset');
        $('#createModal .alert').html('<strong>Error</strong> ' +
          // (status === 302 ? "Already exists" : "Internal server error"));
          (arg.responseText ? arg.responseText : commErrorText));
        $('#createModal .alert').show();
      });
    });

    $('#renameModal').on('show.bs.modal', function () {
      $('#renameModal .alert').hide();

      var src = $('input#modalData').val();
      $('#renameModal .modal-title').text("Rename " + src);
      $('#renameModal input').val(src);
    });

    $('#renameModal').on('shown.bs.modal', function () {
      $('#renameModal input').focus();
    });

    $('#renameModal input').on('keypress', function (ev) {
      if (ev.keyCode === 13)
        $('#renameModalButton').trigger('click');
    })

    $('#renameModalButton').on('click', function () {
      var src = $('input#modalData').val();
      var dst = $('#renameModal input').val();
      if (dst.length === 0) { return; }
      if (dst[0] !== '/') {
        dst = syaro.wikiPath + '/' + dst;
      }

      $('#renameModalButton').button('loading')

      Promise.resolve({
        url: '/api/rename?src=' + encodeURIComponent(src) + ';dst=' + encodeURIComponent(dst),
        method: 'GET',
      })
      .then(xhr)
      .then(function (status) {
        // refreash Page
        location.reload();
      })
      .catch(function (status) {
        $('#renameModalButton').button('reset');
        $('#renameModal .alert').html('<strong>Error</strong> ' +
            (status === 404 ? "Not found" : "Internal server error"));
        $('#renameModal .alert').show();
      });
    });

    $('#deleteModal').on('show.bs.modal', function () {
      $('#deleteModal .alert').hide();

      var wpath = $('input#modalData').val();
      $('#deleteModal .modal-title').text("Delete " + wpath);
      var msg = "<p>Are you sure you want to <strong>DELETE</strong> " + wpath + "?";
      if (syaro.isDir) {
        msg += "<p>All contents under " + wpath + " will be deleted!</p>";
      }
      $('#deleteModal .msg').html(msg);
    });

    $('#deleteModalButton').on('click', function () {
      var wpath = $('input#modalData').val();

      $('#deleteModalButton').button('loading')

      Promise.resolve({
        url: '/api/delete?wpath=' + encodeURIComponent(wpath),
        method: 'GET',
      }).then(xhr)
      .then(function (status) {
        // refreash Page
        location.reload();
      })
      .catch(function (status) {
        $('#deleteModalButton').button('reset');
        $('#deleteModal .alert').html('<strong>Error</strong> ' +
            (status === 404 ? "Not found" : "Internal server error"));
        $('#deleteModal .alert').show();
      });
    });
  }

  function initFileList() {
    $('.uploader-wrapper').hide();
    $('.uploader-wrapper .progress').hide();
    $('.uploader-wrapper .alert').hide();

    // select file list item by clicking and show extra menu
    $('.syaro-filelist table tr').on('click', function (ev) {
      var selected = $(this).hasClass('selected');
      $('.syaro-filelist table tr.selected').removeClass('selected');
      if (!selected) {
        $(this).addClass('selected');
      }

      if ($('.syaro-filelist table tr.selected').length === 0) {
        $('.panel-heading .menu-group li.extra').hide();
      } else {
        $('.panel-heading .menu-group li.extra').show();
      }
    });

    $('.syaro-filelist table').on('dragover', function (ev) {
      ev.preventDefault();
      $('.uploader-wrapper').show();
    });

    $('.uploader-wrapper').on('dragover', function (ev) {
      ev.preventDefault();
      ev.stopPropagation();
      $('.uploader-wrapper').show();
      return false;
    });

    $('.uploader-wrapper').on('dragenter', function (ev) {
      ev.preventDefault();
      ev.stopPropagation();
      return false;
    });

    $('.uploader-wrapper').on('drop', function (ev) {
      ev.preventDefault();
      $('.uploader-wrapper').show();
      $('.uploader-wrapper .uploader-message').hide();
      uploadMulti(ev.originalEvent.dataTransfer.files);
    });

    $('#uploadForm').on('change', function (ev) {
      $('.uploader-wrapper').show();
      $('.uploader-wrapper .uploader-message').hide();
      uploadMulti(this.files);
    });
  }

  function initFileListMenu() {
    $('.panel-heading .menu-group li.extra').hide();

    $('#createPage').on('click', function (ev) {
      ev.preventDefault();
      $('#createModal').modal('show');
    });

    $('#uploadFile').on('mouseover', function () {
      $('.syaro-filelist .uploader-wrapper').show();
    });

    $('#uploadFile').on('mouseout', function () {
      $('.syaro-filelist .uploader-wrapper').hide();
    });

    $('#uploadFile').on('click', function (ev) {
      ev.preventDefault();
      $('#uploadForm').trigger('click');
    });

    $('#renameFile').on('click', function (ev) {
      ev.preventDefault();

      var src = $('.syaro-filelist table tr.selected')[0].children[2].innerText;
      $('input#modalData').val(src);

      $('#renameModal').modal('show');
    });

    $('#deleteFile').on('click', function (ev) {
      ev.preventDefault();

      var wpath = $('.syaro-filelist table tr.selected')[0].children[2].innerText;
      $('input#modalData').val(wpath);

      $('#deleteModal').modal('show');
    });
  }

  // http://www.html5rocks.com/ja/tutorials/file/dndfiles/
  function uploadMulti(files) {
    $('.uploader-wrapper .progress').show();

    var l = files.length, c = 0;
    var ps = [];
    for (var i = 0; i < l; i++) {
      var f = files[i];
      var wpath = syaro.wikiPath + '/' + f.name;

      ps.push(
        (new Promise(function(resolve, reject) {
          var r = new FileReader();
          r.onloadend = function () {
            resolve({
              url: '/api/upload?wpath=' + encodeURIComponent(wpath),
              method: 'POST',
              wpath: wpath,
              body: r.result
            });
          };
          r.onerror = function () {
            reject(r.error);
          }
          r.readAsBinaryString(f);
        }))
        .then(xhr)
        .then(function (arg) {
          c++;
          var now = 100*c/l;
          $('.uploader-wrapper .progress-bar').css({width: now+'%'});
          return arg;
        })
      );
    }

    Promise.all(ps)
    .then(function (arg) {
      location.reload();
    })
    .catch(function (err) {
      $('.uploader-wrapper .progress').hide();
      $('.uploader-wrapper .alert').html('<strong>Error</strong> ' + err.responseText ? err.responseText : err);
      $('.uploader-wrapper .alert').show();
    });
  }

  function scrollNav() {
    var nav = $('.syaro-right nav');
    // get top offset including margin
    // var navTop = nav.position().top + nav.parent().offset().top;
    var navTop = nav.parent().offset().top;
    var scrollTop = $(window).scrollTop();
    if (scrollTop > navTop) {
      nav.addClass('affix');
    } else {
      nav.removeClass('affix');
    }
  }

  function initEmojify() {
    emojify.setConfig({
        mode: 'img',
        img_dir: '/images/emoji',
        ignore_emoticons: true,
    });
    $(".markdown").each(function() {
      emojify.run($(this).get(0));
    });
  }

  function initHistory() {
    $('a[href="#syaro-history"]').on('shown.bs.tab', function () {
      if (historyLoaded) { return; }

      Promise.resolve({
        url: '/api/history?wpath=' + encodeURIComponent(syaro.wikiPath),
        method: 'GET',
      }).then(xhr)
        .then(function (arg) {
          var data = JSON.parse(arg.response);
          $('#syaro-history .progress').hide();

          var panel = '<div class="panel panel-info">' +
            '<div class="panel-heading">History of ' +
            syaro.wikiPath + '</div>'
          var thead = '<thead><tr><th>Op</th><th>Message</th><th>Author</th><th>Date</th></tr></thead>';
          var table = '<table class="table table-striped  table-bordered">' + thead + '<tbody>';
          table += '<thead>'
          table += data.map(function (c) {
            return '<tr><td>' + [c.op, c.msg, c.name, c.date].join('</td><td>') +
              '</td></tr>';
          }).join('');
          table += '</tbody></table>';

          $('#syaro-history').html(panel + table + '</div>');
          historyLoaded = true;
        })
        .catch(function (arg) {
          $('#syaro-history .progress').hide();
          $('#syaro-history .alert').html('<strong>Error</strong> ' +
            arg.status + ' ' + (arg.responseText ? arg.responseText : arg.statusText));
          $('#syaro-history .alert').show();
          historyLoaded = true;
        });
    });
  }

  function xhr(arg) {
    return new Promise(function(resolve, reject) {
      var req = new XMLHttpRequest();
      req.open(arg.method, arg.url);

      req.onreadystatechange = function () {
        if (req.readyState !== XMLHttpRequest.DONE) { return; }

        var arg = $.extend(arg, {
          status: req.status,
          statusText: req.statusText,
          response: req.response,
          responseText: req.responseText,
        });
        if (Math.floor(req.status/100) === 2) {
          resolve(arg);
        } else {
          reject(arg);
        }
      };

      req.onerror = function () {
        reject(arg);
      }

      if (arg.body) req.send(arg.body);
      else req.send();
    });
  }

  init()

})
