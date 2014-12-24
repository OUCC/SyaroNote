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
                    var tableInfo  = getTable(session, cursorPos, 0, session.getLength());
                    var formatInfo = format(tableInfo.table, newLine);
                    session.replace(tableInfo.range, formatInfo.text);

                    var newCursorRow, newCursorColumn;
                    if (tableInfo.focusPos.row === 0) {
                        newCursorRow    = tableInfo.range.start.row + tableInfo.focusPos.row;
                        if (tableInfo.focusPos.column < formatInfo.numColumns) {
                            newCursorColumn = tableInfo.focusPos.column + 2;
                            for (var i = 0; i <= tableInfo.focusPos.column; i++) {
                                newCursorColumn += formatInfo.columnWidth[i] + 2;
                            }
                            editor.clearSelection();
                            editor.moveCursorTo(newCursorRow, newCursorColumn);
                        }
                    }
                    else if (tableInfo.focusPos.column < formatInfo.numColumns - 1) {
                        newCursorRow    = tableInfo.range.start.row + tableInfo.focusPos.row;
                        newCursorColumn = tableInfo.focusPos.column + 2;
                        for (var i = 0; i <= tableInfo.focusPos.column; i++) {
                            newCursorColumn += formatInfo.columnWidth[i] + 2;
                        }
                        editor.clearSelection();
                        editor.moveCursorTo(newCursorRow, newCursorColumn);
                    }
                    else {
                        newCursorRow    = tableInfo.range.start.row + tableInfo.focusPos.row + 1;
                        newCursorColumn = 1;
                        if (session.getLine(newCursorRow)[0] !== "|") {
                            session.insert({ row: newCursorRow - 1, column: session.getLine(newCursorRow - 1).length },
                                newLine + "|");
                        }
                        editor.clearSelection();
                        editor.moveCursorTo(newCursorRow, newCursorColumn);
                    }
                    // avoid insert \t or 4 space
                    return true;
                }
                else {
                    // allow other ace commands to handle event
                    return false;
                }
            },
        }]);
    }

    function getTable(session, cursorPos, minLineNum, maxLineNum) {
        var lines = [];

        var lineNum, line, cleanedLine;

        var currentLineNum   = cursorPos.row;
        var tableHeadLineNum = currentLineNum;

        for (lineNum = currentLineNum - 1; lineNum >= minLineNum; lineNum--) {
            line        = session.getLine(lineNum);
            cleanedLine = removeSpaces(line);
            if (cleanedLine[0] === "|") {
                if (cleanedLine.length > 1 && cleanedLine[cleanedLine.length - 1] === "|") {
                    lines.push(cleanedLine.substring(1, cleanedLine.length - 1));
                }
                else {
                    lines.push(cleanedLine.substring(1));
                }
                tableHeadLineNum = lineNum;
            }
            else {
                break;
            }
        }

        lines.reverse();

        var tableFootLineNum    = currentLineNum;
        var tableFootLineLength = 0;

        for (lineNum = currentLineNum; lineNum <= maxLineNum; lineNum++) {
            line        = session.getLine(lineNum);
            cleanedLine = removeSpaces(line);
            if (cleanedLine[0] === "|") {
                if (cleanedLine.length > 1 && cleanedLine[cleanedLine.length - 1] === "|") {
                    lines.push(cleanedLine.substring(1, cleanedLine.length - 1));
                }
                else {
                    lines.push(cleanedLine.substring(1));
                }
                tableFootLineNum    = lineNum;
                tableFootLineLength = line.length;
            }
            else {
                break;
            }
        }

        return {
            table   : lines.map(function (line) { return line.split("|"); }),
            range   : new Range(tableHeadLineNum, 0, tableFootLineNum, tableFootLineLength),
            focusPos: {
                row   : currentLineNum - tableHeadLineNum,
                column: session.getLine(currentLineNum).substring(0, cursorPos.column).split("|").length - 2
            }
        };
    }

    function removeSpaces(line) {
        return line.split("|").map(function (cell) { return cell.replace(/^\s*(.*?)\s*$/, "$1"); }).join("|");
    }

    var TableAlign = Object.freeze({
        LEFT  : "left",
        RIGHT : "right",
        CENTER: "center",
        DEFAULT: "default",
    });

    function format(table, newLine) {
        var maxRowNum    = table.length - 1;
        var maxColumnNum = 0;

        var rowNum, row, columnNum, cell;

        var pipeRowNum  = -1;
        // compute maxColumnNum and pipeRowNum
        for (rowNum = 0; rowNum <= maxRowNum; rowNum++) {
            row          = table[rowNum];
            maxColumnNum = Math.max(maxColumnNum, row.length - 1);
            if (pipeRowNum < 0 && row.every(isPipeCell)) {
                pipeRowNum = rowNum;
            }
        }

        var columnWidth = [];
        var columnAlign = [];
        // initialize columWidth and columnAlign
        for (columnNum = 0; columnNum <= maxColumnNum; columnNum++){
            columnWidth[columnNum] = 0;
            columnAlign[columnNum] = TableAlign.DEFAULT;
        }
        // compute columnWidth
        for (rowNum = 0; rowNum <= maxRowNum; rowNum++) {
            row = table[rowNum];
            if (rowNum !== pipeRowNum) {
                for (columnNum = 0; columnNum < row.length; columnNum++) {
                    cell = row[columnNum];
                    columnWidth[columnNum] = Math.max(columnWidth[columnNum], cell.length);
                }
            }
        }
        // compute columnAlign
        if (pipeRowNum >= 0) {
            row = table[pipeRowNum];
            for (columnNum = 0; columnNum < row.length; columnNum++) {
                cell = row[columnNum];
                var alignLeft  = cell[0] === ":";
                var alignRight = cell[cell.length - 1] === ":";
                if (alignLeft && alignRight) {
                    columnAlign[columnNum] = TableAlign.CENTER;
                }
                else if (alignLeft) {
                    columnAlign[columnNum] = TableAlign.LEFT;
                }
                else if (alignRight) {
                    columnAlign[columnNum] = TableAlign.RIGHT;
                }
                else {
                    columnAlign[columnNum] = TableAlign.DEFAULT;
                }
            }
        }

        var formattedRowStrs = [];
        // make formatted table string
        for (rowNum = 0; rowNum <= maxRowNum; rowNum++) {
            row = table[rowNum];
            var formattedRowStr = "|";
            if (rowNum === pipeRowNum) {
                for (columnNum = 0; columnNum <= maxColumnNum; columnNum++) {
                    formattedRowStr += formatPipeCell(columnAlign[columnNum], columnWidth[columnNum]) + "|";
                }
            }
            else {
                for (columnNum = 0; columnNum <= maxColumnNum; columnNum++) {
                    formattedRowStr += " "
                        + formatCell(columnAlign[columnNum], columnWidth[columnNum], row[columnNum]) + " |";
                }
            }
            formattedRowStrs.push(formattedRowStr);
        }

        return {
            text       : formattedRowStrs.join(newLine),
            numRows    : maxRowNum + 1,
            numColumns : maxColumnNum + 1,
            columnAlign: columnAlign.slice(),
            columnWidth: columnWidth.slice()
        };
    }

    function isPipeCell(cell) {
        for (var i = 0; i < cell.length; i++) {
            if (cell[i] !== "-" && cell[i] !== ":") {
                return false;
            }
        }
        return true;
    }

    function formatPipeCell(align, width) {
        switch (align) {
            case TableAlign.LEFT:
                return ":" + repeatStr(width, "-") + " ";
            case TableAlign.RIGHT:
                return " " + repeatStr(width, "-") + ":";
            case TableAlign.CENTER:
                return ":" + repeatStr(width, "-") + ":";
            case TableAlign.DEFAULT:
            default:
                return " " + repeatStr(width, "-") + " ";
        }
    }

    function formatCell(align, width, cell) {
        if (cell === undefined) {
            return repeatStr(width, " ");
        }
        else {
            return alignStr(align, width, cell);
        }
    }

    function repeatStr(n, str) {
        var result = "";
        for (var i = 0; i < n; i++) {
            result += str;
        }
        return result;
    }

    function alignStr(align, width, str) {
        if (str.length > width) {
            return str;
        }
        else {
            var spaceWidth = width - str.length;
            switch (align) {
                case TableAlign.RIGHT:
                    return repeatStr(spaceWidth, " ") + str;
                case TableAlign.CENTER:
                    return repeatStr(Math.floor((spaceWidth) / 2), " ") + str
                        + repeatStr(Math.ceil((spaceWidth) / 2), " ");
                case TableAlign.LEFT:
                case TableAlign.DEFAULT:
                default:
                    return str + repeatStr(spaceWidth, " ");
            }
        }
    }

    // export
    global["TableFormatter"] = TableFormatter;

})((this || 0).self || global);
