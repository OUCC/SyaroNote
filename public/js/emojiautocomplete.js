(function (global) {
    "use strict;"

    var HashHandler = ace.require("ace/keyboard/hash_handler").HashHandler;
    var Range = ace.require("ace/range").Range;

    var emojiNames = emojify.emojiNames;
    var $suggestDropdown;
    var enteringEmoji;

    function EmojiAutoComplete(editor) {
        $suggestDropdown = $(_dropdownDom());
        $("#app_wrap").append($suggestDropdown);
        editor.getSession().on("changeBackMarker", function() {
            eacExecEvent(editor);
        });
    }

    function eacExecEvent(editor) {
        var d = new $.Deferred;
        var maybe = detectSuggestableEmoji(editor);

        if (maybe === null) {
            $suggestDropdown.removeClass("open");
        } else {
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
                .css("left", $(".ace_cursor").offset().left      + "px");
            $suggestDropdown.addClass("open");
        }
        d.resolve();
        return d.promise();
    }

    function detectSuggestableEmoji(editor) {
        var session     = editor.getSession();
        var cursorPos   = editor.getCursorPosition();
        var currentLine = session.getLine(cursorPos.row);

        var headMatch = currentLine.slice(0, cursorPos.column).match(/^:([^:\s]*)$|\s:([^:\s]*)$/);
        var tailMatch = currentLine.slice(cursorPos.column).match(/^([^:\s]*)$|^([^:\s]*)\s/);
        if (headMatch === null || tailMatch === null) {
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

    function execHashHandler() {
        return function(editor) {
            console.log("event");
        }
    }

    function _dropdownDom() {
        var dom = '<div class="dropdown emoji-suggest"><ul class="dropdown-menu emoji-suggest-list">';
        emojiNames.forEach(function(e) {
            dom += '<li><a class="insert-emoji-trigger hidden" data-emoji="' + e + '">'
                + emojify.replace(':' + e + ':') + '<span class="dropdown-emoji-text"></span>'
                + '</a></li>';
        });
        dom += '</ul></div>';
        return dom;
    }
    // export
    global["EmojiAutoComplete"] = EmojiAutoComplete;

})((this || 0).self || global);
