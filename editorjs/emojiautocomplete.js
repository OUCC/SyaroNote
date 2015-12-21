import emojify from 'emojify.js'

var HashHandler = ace.require("ace/keyboard/hash_handler").HashHandler;
var Range = ace.require("ace/range").Range;

var emojiNames = emojify.emojiNames;
var $suggestDropdown;
var enteringEmoji;
var maybe = [];

export default function EmojiAutoComplete(editor) {
    $suggestDropdown = $(_dropdownDom());
    $("#app_wrap").append($suggestDropdown);
    $suggestDropdown.find(".insert-emoji-trigger").on("click", function() {
        insertEmoji(editor, $(this).attr("data-emoji"));
    });

    editor.getSession().on("changeBackMarker", function() {
        cursorMoveAction(editor);
    });
    editor.keyBinding.addKeyboardHandler(new HashHandler([{
        bindKey: "Tab",
        descr:   "Insert suggested emoji",
        exec:    tabKeyAction(editor)
    }]));
}

function cursorMoveAction(editor) {
    var m = detectSuggestableEmoji(editor);
    if (m === null) {
        hideSuggest();
    } else {
        showSuggest(m);
    }
}

function tabKeyAction(editor) {
    return function () {
        if (maybe.length > 0) {
            insertEmoji(editor, maybe[0]);
            hideSuggest();
            return true;
        }
        return false;
    };
}

function detectSuggestableEmoji(editor) {
    var session     = editor.getSession();
    var cursorPos   = editor.getCursorPosition();
    var currentLine = session.getLine(cursorPos.row);

    var headMatch = currentLine.slice(0, cursorPos.column).match(/^:([\w+-]*)$|\s:([\w+-]*)$/);
    var tailMatch = currentLine.slice(cursorPos.column).match(/^[\w+-]*:/);
    if (headMatch === null || tailMatch !== null) {
        return null;
    }

    enteringEmoji = (typeof headMatch[1] !== "undefined")? headMatch[1] : headMatch[2];
    var maybe = emojiNames.filter(function(e) {
        return (e.indexOf(enteringEmoji) === 0);
    });
    if (maybe.length === 0) {
        return null;
    }

    return maybe;
}

function insertEmoji(editor, emoji) {
    var session     = editor.getSession();
    var cursorPos   = editor.getCursorPosition();
    var currentLine = session.getLine(cursorPos.row);

    var headMatch = currentLine.slice(0, cursorPos.column).match(/^(.*:)[\w+-:w]*$/);
    var tailMatch = currentLine.slice(cursorPos.column).match(/^([\w+-]*)\s|^([\w+-]*)/);
    if (headMatch === null || tailMatch === null) {
        return null;
    }

    var startColumn = headMatch[1].length;
    var endColumn = cursorPos.column
                  + ((typeof tailMatch[1] !== "undefined")? tailMatch[1].length : tailMatch[2].length);
    var text = emoji + ((typeof tailMatch[1] !== "undefined")? ":" : ": ");

    session.replace(new Range(cursorPos.row, startColumn, cursorPos.row, endColumn), text);
    editor.moveCursorTo(cursorPos.row, startColumn + text.length);
    editor.focus();
}

function showSuggest(m) {
    maybe = m;
    $suggestDropdown.find(".insert-emoji-trigger").each(function() {
        if (enteringEmoji === "") {
            $(this).removeClass("hidden")
                .children(".dropdown-emoji-text").text($(this).attr("data-emoji"));
        } else {
            if (maybe.indexOf($(this).attr("data-emoji")) >= 0) {
                $(this).removeClass("hidden")
                    .children(".dropdown-emoji-text").html(
                        '<b>' + enteringEmoji + '</b>' + $(this).attr("data-emoji").slice(enteringEmoji.length)
                    );
            } else {
                $(this).addClass("hidden");
            }
        }
    });
    $suggestDropdown.find("ul.emoji-suggest-list")
        .css("top", ($(".ace_cursor").offset().top + 40) + "px")
        .css("left", $(".ace_cursor").offset().left      + "px")
        .scrollTop(0);
    $suggestDropdown.addClass("open");
}

function hideSuggest() {
    maybe = [];
    $suggestDropdown.removeClass("open");
}

function _dropdownDom() {
    var dom = '<div class="dropdown emoji-suggest"><ul class="dropdown-menu emoji-suggest-list">';
    emojiNames.forEach(function(e) {
        dom += '<li><a class="insert-emoji-trigger hidden" data-emoji="' + e + '">'
            + '<span class="emoji emoji-' + e + '" title=":' + e + ':"></span>'
            + '<span class="dropdown-emoji-text"></span>'
            + '</a></li>';
    });
    dom += '</ul></div>';
    return dom;
}
