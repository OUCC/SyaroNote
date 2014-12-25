(function (global) {
    "use strict;"

    var HashHandler = ace.require("ace/keyboard/hash_handler").HashHandler;
    var Range = ace.require("ace/range").Range;

    function TableFormatter() {
        return new HashHandler([{
            bindKey: "Tab",
            descr  : "Format markdown table",
            exec   : function (editor) {
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
                    moveCursor(editor, session, newTableInfo, newLine);
                    // avoid insert \t or 4 space
                    return true;
                }
                else {
                    // allow other ace commands to handle event
                    return false;
                }
            }
        }]);
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

    function moveCursor(editor, session, tableInfo, newLine) {
        var focusPos = tableInfo.focusPos;
        var nextCell;
        var newCursorRow, newCursorColumn;
        if (focusPos.column < tableInfo.numColumns - 1) {
            newCursorRow    = tableInfo.range.start.row + focusPos.row;
            nextCell        = tableInfo.table[focusPos.row][focusPos.column + 1];
            if (nextCell === undefined || nextCell.cleaned === "") {
                newCursorColumn = focusPos.column + 3;
            }
            else {
                newCursorColumn = focusPos.column + 2
                    + nextCell.raw.indexOf(nextCell.cleaned) + nextCell.cleaned.length;
            }
            for (var i = 0; i <= focusPos.column; i++) {
                newCursorColumn += tableInfo.columnWidth[i] + 2;
            }
        }
        else if (focusPos.row === 0 && focusPos.column == tableInfo.numColumns - 1) {
            newCursorRow = tableInfo.range.start.row + focusPos.row;
            newCursorColumn = focusPos.column + 3;
            for (var i = 0; i <= focusPos.column; i++) {
                newCursorColumn += tableInfo.columnWidth[i] + 2;
            }
            session.insert(
                { row: newCursorRow, column: session.getLine(newCursorRow).length },
                " "
            );
        }
        else {
            var nextRowNum;
            if (focusPos.row === 0 && focusPos.row + 1 === tableInfo.alignRowNum) {
                nextRowNum = focusPos.row + 2;
            }
            else {
                nextRowNum = focusPos.row + 1;
            }
            newCursorRow = tableInfo.range.start.row + nextRowNum;
            if (nextRowNum > tableInfo.numRows - 1) {
                newCursorColumn = 2;
                session.insert(
                    { row: newCursorRow - 1, column: session.getLine(newCursorRow - 1).length },
                    newLine + "| "
                );
            }
            else {
                nextCell = tableInfo.table[nextRowNum][0];
                if (nextCell === undefined || nextCell.cleaned === "") {
                    newCursorColumn = 2;
                }
                else {
                    newCursorColumn = 1 + nextCell.raw.indexOf(nextCell.cleaned) + nextCell.cleaned.length;
                }
            }
        }
        editor.clearSelection();
        editor.moveCursorTo(newCursorRow, newCursorColumn);
    }

    // export
    global["TableFormatter"] = TableFormatter;

})((this || 0).self || global);
