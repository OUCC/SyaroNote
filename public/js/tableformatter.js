(function (global) {
    "use strict;"

    var HashHandler = ace.require("ace/keyboard/hash_handler").HashHandler;
    var Range = ace.require("ace/range").Range;

    function TableFormatter() {
        return new HashHandler([
            {
                bindKey: "Tab",
                descr  : "Format markdown table and move cursor to next cell",
                exec   : execHashHandler(true)
            },
            {
                bindKey: "Shift+Tab",
                descr  : "Format markdown table and move cursor to previous cell",
                exec   : execHashHandler(false)
            }
        ]);
    }

    function execHashHandler(forward) {
        return function(editor) {
            var session     = editor.getSession();
            var cursorPos   = editor.getCursorPosition();
            var currentLine = session.getLine(cursorPos.row);
            var newLine     = session.getNewLineMode() === "windows" ? "\r\n" : "\n";
            if (currentLine[0] === "|") {
                var tableInfo     = getTableInfo(session, cursorPos, 0, session.getLength());
                var formattedText = format(tableInfo, newLine);
                session.replace(tableInfo.range, formattedText);
                var newTableInfo = getTableInfo(
                    session, { row: tableInfo.range.start.row, column: 0}, 0, session.getLength()
                );
                newTableInfo.focusPos = tableInfo.focusPos;
                moveCursor(editor, session, newTableInfo, newLine, forward);
                // avoid insert \t or 4 space
                return true;
            }
            else {
                // allow other ace commands to handle event
                return false;
            }
        }
    }

    var CellAlign = Object.freeze({
        LEFT  : "left",
        RIGHT : "right",
        CENTER: "center",
        DEFAULT: "default"
    });

    function getTableInfo(session, cursorPos, minLineNum, maxLineNum) {
        var table = [];

        var lineNum, line, row;

        var currentLineNum      = cursorPos.row;
        var tableHeadLineNum    = currentLineNum;
        var tableFootLineNum    = currentLineNum;
        var tableFootLineLength = 0;

        // get table head
        for (lineNum = currentLineNum - 1; lineNum >= minLineNum; lineNum--) {
            line = session.getLine(lineNum);
            if (line[0] === "|") {
                row = line.split("|").map(function (cell) {
                    return {
                        raw    : cell,
                        cleaned: cell.replace(/^\s*(.*?)\s*$/, "$1")
                    };
                });
                if (row[row.length - 1].cleaned === "") {
                    table.push(row.slice(1, row.length - 1));
                }
                else {
                    table.push(row.slice(1));
                }
                tableHeadLineNum = lineNum;
            }
            else {
                break;
            }
        }

        table.reverse();

        // get table foot
        for (lineNum = currentLineNum; lineNum <= maxLineNum; lineNum++) {
            line = session.getLine(lineNum);
            if (line[0] === "|") {
                row = line.split("|").map(function (rawCell) {
                    return {
                        raw    : rawCell,
                        cleaned: rawCell.replace(/^\s*(.*?)\s*$/, "$1")
                    };
                });
                if (row[row.length - 1].cleaned === "") {
                    table.push(row.slice(1, row.length - 1));
                }
                else {
                    table.push(row.slice(1));
                }
                tableFootLineNum    = lineNum;
                tableFootLineLength = line.length;
            }
            else {
                break;
            }
        }

        var numRows    = table.length;
        var numColumns = 0;

        var rowNum, row, columnNum, cell;

        var alignRowNum  = -1;

        // compute numColumns and alignRowNum
        for (rowNum = 0; rowNum < numRows; rowNum++) {
            row        = table[rowNum];
            numColumns = Math.max(numColumns, row.length);
            if (rowNum === 1 && alignRowNum < 0 && row.every(isAlignCell)) {
                alignRowNum = rowNum;
            }
        }

        var columnWidth = [];
        var columnAlign = [];
        // initialize columWidth and columnAlign
        for (columnNum = 0; columnNum < numColumns; columnNum++){
            columnWidth[columnNum] = 0;
            columnAlign[columnNum] = CellAlign.DEFAULT;
        }
        // compute columnWidth
        for (rowNum = 0; rowNum < numRows; rowNum++) {
            if (rowNum !== alignRowNum) {
                row = table[rowNum];
                for (columnNum = 0; columnNum < row.length; columnNum++) {
                    cell                   = row[columnNum];
                    columnWidth[columnNum] = Math.max(columnWidth[columnNum], cell.cleaned.length);
                }
            }
        }
        // compute columnAlign
        if (alignRowNum >= 0) {
            row = table[alignRowNum];
            for (columnNum = 0; columnNum < row.length; columnNum++) {
                cell = row[columnNum];
                var alignLeft  = cell.cleaned[0] === ":";
                var alignRight = cell.cleaned[cell.cleaned.length - 1] === ":";
                if (alignLeft && alignRight) {
                    columnAlign[columnNum] = CellAlign.CENTER;
                }
                else if (alignLeft) {
                    columnAlign[columnNum] = CellAlign.LEFT;
                }
                else if (alignRight) {
                    columnAlign[columnNum] = CellAlign.RIGHT;
                }
                else {
                    columnAlign[columnNum] = CellAlign.DEFAULT;
                }
            }
        }

        return {
            table      : table,
            range      : new Range(tableHeadLineNum, 0, tableFootLineNum, tableFootLineLength),
            focusPos   : {
                row   : currentLineNum - tableHeadLineNum,
                column: session.getLine(currentLineNum).substring(0, cursorPos.column).split("|").length - 2
            },
            numRows    : numRows,
            numColumns : numColumns,
            alignRowNum: alignRowNum,
            columnAlign: columnAlign.slice(),
            columnWidth: columnWidth.slice()
        };
    }

    function format(tableInfo, newLine) {
        var table       = tableInfo.table.map(function (row) { return row.slice(); });
        var numRows     = tableInfo.numRows;
        var alignRowNum = tableInfo.alignRowNum;
        if (tableInfo.numRows === 1 && alignRowNum < 0) {
            table.push([]);
            numRows++;
            alignRowNum = 1;
        }
        var rowTexts = [];
        for (var rowNum = 0; rowNum < numRows; rowNum++) {
            row = table[rowNum];
            var columnNum;
            var rowText = "|";
            if (rowNum === alignRowNum) {
                for (columnNum = 0; columnNum < tableInfo.numColumns; columnNum++) {
                    rowText += formatAlignCell(
                            tableInfo.columnAlign[columnNum], tableInfo.columnWidth[columnNum]
                        ) + "|";
                }
            }
            else {
                for (columnNum = 0; columnNum < tableInfo.numColumns; columnNum++) {
                    rowText += " " + formatCell(
                            tableInfo.columnAlign[columnNum], tableInfo.columnWidth[columnNum], row[columnNum]
                        ) + " |";
                }
            }
            rowTexts.push(rowText);
        }
        return rowTexts.join(newLine);
    }

    function isAlignCell(cell) {
        for (var i = 0; i < cell.cleaned.length; i++) {
            if (cell.cleaned[i] !== "-" && cell.cleaned[i] !== ":") {
                return false;
            }
        }
        return true;
    }

    function formatAlignCell(align, width) {
        switch (align) {
            case CellAlign.LEFT:
                return ":" + repeatStr(width, "-") + " ";
            case CellAlign.RIGHT:
                return " " + repeatStr(width, "-") + ":";
            case CellAlign.CENTER:
                return ":" + repeatStr(width, "-") + ":";
            case CellAlign.DEFAULT:
            default:
                return " " + repeatStr(width, "-") + " ";
        }
    }

    function formatCell(align, width, cell) {
        if (cell === undefined) {
            return repeatStr(width, " ");
        }
        else {
            return alignText(align, width, cell.cleaned);
        }
    }

    function repeatStr(n, str) {
        var result = "";
        for (var i = 0; i < n; i++) {
            result += str;
        }
        return result;
    }

    function alignText(align, width, text) {
        if (text.length > width) {
            return text;
        }
        else {
            var spaceWidth = width - text.length;
            switch (align) {
                case CellAlign.RIGHT:
                    return repeatStr(spaceWidth, " ") + text;
                case CellAlign.CENTER:
                    return repeatStr(Math.floor((spaceWidth) / 2), " ") + text
                        + repeatStr(Math.ceil((spaceWidth) / 2), " ");
                case CellAlign.LEFT:
                case CellAlign.DEFAULT:
                default:
                    return text + repeatStr(spaceWidth, " ");
            }
        }
    }

    function moveCursor(editor, session, tableInfo, newLine, forward) {
        if (forward) {
            moveCursorForward(editor, session, tableInfo, newLine);
        }
        else {
            moveCursorBack(editor, session, tableInfo, newLine);
        }
    }

    function moveCursorForward(editor, session, tableInfo, newLine) {
        var focusPos = tableInfo.focusPos;
        if (focusPos.column < tableInfo.numColumns - 1) {
            // move to the next column in the same row
            moveCursorToCell(editor, tableInfo, focusPos.row, focusPos.column + 1);
        }
        else if (focusPos.row === 0 && focusPos.column == tableInfo.numColumns - 1) {
            // move to the last of the header row
            session.insert(
                {
                    row   : tableInfo.range.start.row + focusPos.row,
                    column: session.getLine(tableInfo.range.start.row + focusPos.row).length
                },
                " "
            );
            moveCursorToCell(editor, tableInfo, focusPos.row, focusPos.column + 1);
        }
        else {
            // move to the first column in the lower row
            var nextRowNum;
            if (focusPos.row === 0 && focusPos.row + 1 === tableInfo.alignRowNum) {
                // skip the alignment row
                nextRowNum = focusPos.row + 2;
            }
            else {
                nextRowNum = focusPos.row + 1;
            }

            if (nextRowNum > tableInfo.numRows - 1) {
                session.insert(
                    {
                        row   : tableInfo.range.start.row + nextRowNum - 1,
                        column: session.getLine(tableInfo.range.start.row + nextRowNum - 1).length
                    },
                    newLine + "| "
                );
            }
            moveCursorToCell(editor, tableInfo, nextRowNum, 0);
        }
    }

    function moveCursorBack(editor, session, tableInfo, newLine) {
        var focusPos = tableInfo.focusPos;
        if (focusPos.column > 0) {
            // move to the next column in the same row
            moveCursorToCell(editor, tableInfo, focusPos.row, focusPos.column - 1);
        }
        else if (focusPos.row > 0) {
            // move to the last column in the upper row
            var nextRowNum;
            if (focusPos.row - 1 === tableInfo.alignRowNum) {
                // skip the alignment row
                nextRowNum = focusPos.row - 2;
            }
            else {
                nextRowNum = focusPos.row - 1;
            }
            moveCursorToCell(editor, tableInfo, nextRowNum, tableInfo.numColumns - 1);
        }
        else {
            // don't move
            moveCursorToCell(editor, tableInfo, focusPos.row, focusPos.column);
        }
    }

    function moveCursorToCell(editor, tableInfo, row, column) {
        if (row < 0 || column < 0) {
            return;
        }
        else if (row >= tableInfo.numRows) {
            row    = tableInfo.numRows;
            column = 0;
        }
        else if (column >= tableInfo.numColumns) {
            row    = 0;
            column = tableInfo.numColumns;
        }
        var newCursorRow = tableInfo.range.start.row + row;
        var newCursorColumn;
        var nextCell = tableInfo.table[row] === undefined ? undefined : tableInfo.table[row][column];
        if (nextCell === undefined || nextCell.cleaned === "") {
            newCursorColumn = column + 2;
        }
        else {
            newCursorColumn = column + 1 + nextCell.raw.indexOf(nextCell.cleaned) + nextCell.cleaned.length;
        }
        for (i = 0; i < column; i++) {
            newCursorColumn += tableInfo.columnWidth[i] + 2;
        }
        editor.clearSelection();
        editor.moveCursorTo(newCursorRow, newCursorColumn);
    }

    // export
    global["TableFormatter"] = TableFormatter;

})((this || 0).self || global);
