(function (global) {
    "use strict;"

    var HashHandler = ace.require("ace/keyboard/hash_handler").HashHandler;
    var Range = ace.require("ace/range").Range;

    function TableFormatter() {

        return new HashHandler([{
            bindKey: "Tab",
            descr  : "Format markdown table",
            exec   : function (editor) {
                var session        = editor.getSession();
                var cursorPos      = editor.getCursorPosition();
                var currentLine    = session.getLine(cursorPos.row);
                if (currentLine[0] === "|") {
                    var tableInfo      = getTable(session, cursorPos.row, 0, session.getLength());
                    var formattedTable = format(tableInfo.table);
                    session.replace(tableInfo.range, formattedTable);

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

    function getTable(session, currentLineNum, minLineNum, maxLineNum) {
        var lines = [];

        var lineNum, line, cleanedLine;
        
        var tableHeadLineNum = currentLineNum;

        for (lineNum = currentLineNum - 1; lineNum >= minLineNum; lineNum--) {
            line        = session.getLine(lineNum);
            cleanedLine = removeSpaces(line);
            if (cleanedLine[0] === "|") {
                if (cleanedLine[cleanedLine.length - 1] === "|") {
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
                if (cleanedLine[cleanedLine.length - 1] === "|") {
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
            table: lines.map(function (line) { return line.split("|"); }),
            range: new Range(tableHeadLineNum, 0, tableFootLineNum, tableFootLineLength)
        };
    }

    function removeSpaces(line) {
        var result = "";
        for (var i = 0; i < line.length; i++) {
            if (" \t\n\r\f\v".indexOf(line[i]) < 0) {
                result += line[i];
            }
        }
        return result;
    }

    var TableAlign = Object.freeze({
        LEFT  : "left",
        RIGHT : "right",
        CENTER: "center"
    });

    function format(table) {
        var maxRowNum    = table.length;
        var maxColumnNum = 0;
        
        var rowNum, row, columnNum, cell;

        var pipeRowNum  = -1;
        // compute maxColumnNum and pipeRowNum
        for (rowNum = 0; rowNum < maxRowNum; rowNum++) {
            row          = table[rowNum];
            maxColumnNum = Math.max(maxColumnNum, row.length);
            if (pipeRowNum < 0 && row.every(isPipeCell)) {
                pipeRowNum = rowNum;
            }
        }

        var columnWidth = [];
        var columnAlign = [];
        // initialize columWidth and columnAlign
        for (columnNum = 0; columnNum < maxColumnNum; columnNum++){
            columnWidth[columnNum] = 0;
            columnAlign[columnNum] = TableAlign.CENTER;
        }
        // compute columnWidth
        for (rowNum = 0; rowNum < maxRowNum; rowNum++) {
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
                    columnAlign[columnNum] = TableAlign.CENTER;
                }
            }
        }

        var formattedRowStrs = [];
        // make formatted table string
        for (rowNum = 0; rowNum < maxRowNum; rowNum++) {
            row = table[rowNum];
            var formattedRowStr = "|";
            if (rowNum === pipeRowNum) {
                for (columnNum = 0; columnNum < maxColumnNum; columnNum++) {
                    formattedRowStr += formatPipeCell(columnAlign[columnNum], columnWidth[columnNum]) + "|";
                }
            }
            else {
                for (columnNum = 0; columnNum < maxColumnNum; columnNum++) {
                    formattedRowStr += " "
                        + formatCell(columnAlign[columnNum], columnWidth[columnNum], row[columnNum]) + " |";
                }
            }
            formattedRowStrs.push(formattedRowStr);
        }

        return formattedRowStrs.join("\n");
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
                return ":" + repeatStr(width, "-") + "-";
            case TableAlign.RIGHT:
                return "-" + repeatStr(width, "-") + ":";
            case TableAlign.CENTER:
            default:
                return ":" + repeatStr(width, "-") + ":";
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
                case TableAlign.LEFT:
                    return str + repeatStr(spaceWidth, " ");
                case TableAlign.RIGHT:
                    return repeatStr(spaceWidth, " ") + str;
                case TableAlign.CENTER:
                default:
                    return repeatStr(Math.floor((spaceWidth) / 2), " ") + str
                        + repeatStr(Math.ceil((spaceWidth) / 2), " ");
            }
        }
    }

    // export
    global["TableFormatter"] = TableFormatter;

})((this || 0).self || global);
