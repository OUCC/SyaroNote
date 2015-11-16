var ControlKey;
(function (ControlKey) {
    ControlKey[ControlKey["Space"] = 0] = "Space";
    ControlKey[ControlKey["Left"] = 1] = "Left";
    ControlKey[ControlKey["Right"] = 2] = "Right";
    ControlKey[ControlKey["Up"] = 3] = "Up";
    ControlKey[ControlKey["Down"] = 4] = "Down";
    ControlKey[ControlKey["Backspace"] = 5] = "Backspace";
    ControlKey[ControlKey["Delete"] = 6] = "Delete";
    ControlKey[ControlKey["Tab"] = 7] = "Tab";
    ControlKey[ControlKey["Enter"] = 8] = "Enter";
    ControlKey[ControlKey["Shift"] = 9] = "Shift";
})(ControlKey || (ControlKey = {}));
const keyMapBlink = {
    "U+0031": ["1", "!"],
    "U+0032": ["2", "\""],
    "U+0033": ["3", "#"],
    "U+0034": ["4", "$"],
    "U+0035": ["5", "%"],
    "U+0036": ["6", "&"],
    "U+0037": ["7", "'"],
    "U+0038": ["8", "("],
    "U+0039": ["9", ")"],
    "U+00BD": ["-", "="],
    "U+00DE": ["^", "~"],
    "U+00DC": ["¥", "|"],
    "U+00C0": ["@", "`"],
    "U+00DB": ["[", "{"],
    "U+00BB": [";", "+"],
    "U+00BA": [":", "*"],
    "U+00DD": ["]", "}"],
    "U+00BC": [",", "<"],
    "U+00BE": [".", ">"],
    "U+00BF": ["/", "?"],
    "U+00E2": ["\\", "_"],
};
function knowKey(e) {
    var code = e.keyCode;
    var key = "";
    if (e.key !== undefined && e.key != " " && e.key.length == 1)
        key = e.key;
    else if ("keyIdentifier" in e) {
        var id = e.keyIdentifier;
        if (id in keyMapBlink)
            key = keyMapBlink[id][e.shiftKey ? 1 : 0];
        else {
            key = getAsciiKey(parseInt(id.substr(2), 16));
            if (!e.shiftKey)
                key = key.toLowerCase();
        }
    }
    return key;
}
function getAsciiKey(code) {
    if (code >= 0x21 && code <= 0x7E)
        return String.fromCharCode(code);
    else
        return "";
}
function knowControlKey(e) {
    switch (e.keyCode) {
        case 8: return ControlKey.Backspace;
        case 9: return ControlKey.Tab;
        case 13: return ControlKey.Enter;
        case 16: return ControlKey.Shift;
        case 32: return ControlKey.Space;
        case 37: return ControlKey.Left;
        case 38: return ControlKey.Up;
        case 39: return ControlKey.Right;
        case 40: return ControlKey.Down;
        case 46: return ControlKey.Delete;
    }
    return null;
}

/// <reference path="typings/jquery.d.ts" />
class Token {
    constructor() {
        this.renderedElem = null;
    }
    clone(parent) {
        console.error("[Token.clone] clone method not implemented");
        return null;
    }
}
class Symbol extends Token {
    constructor(str, variable) {
        super();
        this.str = str;
        this.variable = variable;
    }
    clone(parent) {
        return new Symbol(this.str, this.variable);
    }
    toString() {
        return this.str;
    }
}
class Num extends Token {
    constructor(n) {
        super();
        this.value = n;
    }
    clone(parent) {
        return new Num(this.value);
    }
    toString() {
        return this.value;
    }
}
var StructType;
(function (StructType) {
    StructType[StructType["Infer"] = 0] = "Infer";
    StructType[StructType["Frac"] = 1] = "Frac";
    StructType[StructType["Power"] = 2] = "Power";
    StructType[StructType["Index"] = 3] = "Index";
    StructType[StructType["Matrix"] = 4] = "Matrix";
    StructType[StructType["Diagram"] = 5] = "Diagram";
    StructType[StructType["BigOpr"] = 6] = "BigOpr";
    StructType[StructType["Accent"] = 7] = "Accent";
    StructType[StructType["Macro"] = 8] = "Macro";
})(StructType || (StructType = {}));
class Structure extends Token {
    constructor(parent, type, count) {
        super();
        this.prev = (f) => {
            var i = this.elems.indexOf(f);
            if (i < 0 || i == 0)
                return null;
            return this.elems[i - 1];
        };
        this.next = (f) => {
            var i = this.elems.indexOf(f);
            if (i < 0 || i == this.elems.length - 1)
                return null;
            return this.elems[i + 1];
        };
        this.indexOf = (t) => {
            if (t instanceof Formula)
                return this.elems.indexOf(t);
            else
                return -1;
        };
        this.parent = parent;
        this.type = type;
        var n = 0;
        if (count)
            n = count;
        else {
            switch (type) {
                case StructType.Infer:
                    n = 3;
                    break;
                case StructType.Frac:
                case StructType.BigOpr:
                    n = 2;
                    break;
                case StructType.Power:
                case StructType.Index:
                case StructType.Accent:
                    n = 1;
                    break;
            }
        }
        this.elems = new Array(n);
        for (var i = 0; i < n; i++)
            this.elems[i] = new Formula(this);
    }
    count() {
        return this.elems.length;
    }
    token(i) {
        if (i < 0 || i >= this.elems.length)
            console.error("[Structure.token] out of range : " + i);
        return this.elems[i];
    }
    remove(from, to) {
        var i = Math.min(from);
        var j = Math.max(to);
        var r = this.elems.slice(i, j + 1);
        for (var k = i; k <= j; k++)
            this.elems[k] = new Formula(this);
        return r;
    }
    copy(from, to) {
        return this.elems.slice(from, to).map(s => s.clone(null));
    }
    paste(index, tokens) {
        if (tokens.every(t => t instanceof Formula)) {
            for (var i = 0; i < tokens.length; i++)
                this.elems[index + i] = tokens[i].clone(this);
            return index + tokens.length;
        }
        else {
            this.elems[index].paste(this.elems[index].count(), tokens);
            return index;
        }
    }
    clone(parent) {
        var s = new Structure(parent, this.type);
        this.elems.forEach((f, i) => {
            s.elems[i] = f.clone(s);
        });
        return s;
    }
    toString() {
        var type;
        switch (this.type) {
            case StructType.Frac:
                type = "Frac";
                break;
            case StructType.Infer:
                type = "Infer";
                break;
            case StructType.Power:
                type = "Power";
                break;
            case StructType.Index:
                type = "Index";
                break;
            case StructType.BigOpr:
                type = "BigOpr";
                break;
        }
        return type + "[" + this.elems.map(f => f.toString()).join(", ") + "]";
    }
}
class Matrix extends Structure {
    constructor(parent, rows, cols) {
        super(parent, StructType.Matrix, rows * cols);
        this.rows = rows;
        this.cols = cols;
    }
    pos(index) {
        return { row: Math.floor(index / this.cols), col: index % this.cols };
    }
    tokenAt(r, c, value) {
        if (r < 0 || r >= this.rows || c < 0 || c >= this.cols) {
            console.error("[Matrix.tokenAt] out of range : " + r + "," + c);
            return null;
        }
        else if (value !== undefined)
            return this.elems[r * this.cols + c] = value;
        else
            return this.elems[r * this.cols + c];
    }
    remove(from, to) {
        var r = this.getRectIndex(from, to);
        return [this.cloneRect(r.i1, r.j1, r.i2, r.j2, true)];
    }
    copy(from, to) {
        var r = this.getRectIndex(from, to);
        return [this.cloneRect(r.i1, r.j1, r.i2, r.j2, false)];
    }
    getRectIndex(from, to) {
        var a = this.pos(from);
        var b = this.pos(to);
        var i1 = Math.min(a.row, b.row);
        var j1 = Math.min(a.col, b.col);
        var i2 = Math.max(a.row, b.row);
        var j2 = Math.max(a.col, b.col);
        return { i1: i1, j1: j1, i2: i2, j2: j2 };
    }
    cloneRect(i1, j1, i2, j2, erase) {
        var m = new Matrix(null, i2 - i1 + 1, j2 - j1 + 1);
        for (var i = 0; i < m.rows; i++)
            for (var j = 0; j < m.cols; j++) {
                m.tokenAt(i, j, this.tokenAt(i + i1, j + j1).clone(m));
                if (erase)
                    this.tokenAt(i + i1, j + j1, new Formula(this));
            }
        return m;
    }
    paste(index, tokens) {
        if (tokens.length != 1 || !(tokens[0] instanceof Matrix))
            return super.paste(index, tokens);
        var m = tokens[0];
        var p = this.pos(index);
        var r = Math.min(m.rows, this.rows - p.row);
        var c = Math.min(m.cols, this.cols - p.col);
        for (var i = 0; i < r; i++)
            for (var j = 0; j < c; j++)
                this.tokenAt(p.row + i, p.col + j, m.tokenAt(i, j).clone(this));
        return index + (r - 1) * this.cols + (c - 1);
    }
    clone(parent) {
        var m = new Matrix(parent, this.rows, this.cols);
        this.elems.forEach((f, i) => {
            m.elems[i] = f.clone(m);
        });
        return m;
    }
    extend(horizontal) {
        if (horizontal) {
            for (var i = this.rows; i >= 1; i--)
                this.elems.splice(this.cols * i, 0, new Formula(this));
            this.cols++;
        }
        else {
            for (var i = 0; i < this.cols; i++)
                this.elems.push(new Formula(this));
            this.rows++;
        }
    }
    shrink(horizontal) {
        if (horizontal) {
            if (this.cols == 1)
                return;
            for (var i = this.cols; i >= 1; i--)
                this.elems.splice(this.cols * i - 1, 1);
            this.cols--;
        }
        else {
            if (this.rows == 1)
                return;
            this.elems.splice((this.rows - 1) * this.cols, this.cols);
            this.rows--;
        }
    }
    around(f, horizontal, forward) {
        var i = this.elems.indexOf(f);
        if (i < 0)
            return null;
        var r = Math.floor(i / this.cols);
        var c = i % this.cols;
        if (horizontal) {
            if (!forward && --c < 0)
                return null;
            if (forward && ++c >= this.cols)
                return null;
        }
        else {
            if (!forward && --r < 0)
                return null;
            if (forward && ++r >= this.rows)
                return null;
        }
        return this.tokenAt(r, c);
    }
    nonEmpty(i0, j0, rows, cols) {
        for (var i = 0; i < rows; i++)
            for (var j = 0; j < cols; j++) {
                if (this.tokenAt(i0 + i, j0 + j).tokens.length > 0)
                    return true;
            }
        return false;
    }
    toString() {
        return "Matrix" + this.rows + "," + this.cols + "[" + this.elems.map(f => f.toString()).join(", ") + "]";
    }
}
class BigOpr extends Structure {
    constructor(parent, operator) {
        super(parent, StructType.BigOpr);
        this.operator = operator;
    }
    clone(parent) {
        var s = new BigOpr(parent, this.operator);
        this.elems.forEach((f, i) => {
            s.elems[i] = f.clone(s);
        });
        return s;
    }
}
class Accent extends Structure {
    constructor(parent, symbol, above) {
        super(parent, StructType.Accent);
        this.symbol = symbol;
        this.above = above;
    }
    clone(parent) {
        var s = new Accent(parent, this.symbol, this.above);
        this.elems.forEach((f, i) => {
            s.elems[i] = f.clone(s);
        });
        return s;
    }
    toString() {
        return "Accent" + this.symbol + "[" + this.elems.map(f => f.toString()).join(", ") + "]";
    }
}
class Macro extends Structure {
    constructor(parent, name, argc) {
        super(parent, StructType.Macro, argc);
        this.name = name;
    }
    clone(parent) {
        var m = new Macro(parent, this.name, this.elems.length);
        this.elems.forEach((f, i) => {
            m.elems[i] = f.clone(m);
        });
        return m;
    }
    toString() {
        var str = "\\" + this.name;
        if (this.elems.length > 0) {
            var arg = this.elems.map(e => e.toString()).join("}{");
            str += "{" + arg + "}";
        }
        return str;
    }
}
var FontStyle;
(function (FontStyle) {
    FontStyle[FontStyle["Normal"] = 0] = "Normal";
    FontStyle[FontStyle["Bold"] = 1] = "Bold";
    FontStyle[FontStyle["Roman"] = 2] = "Roman";
    FontStyle[FontStyle["Script"] = 3] = "Script";
    FontStyle[FontStyle["Fraktur"] = 4] = "Fraktur";
    FontStyle[FontStyle["BlackBoard"] = 5] = "BlackBoard";
    FontStyle[FontStyle["Typewriter"] = 6] = "Typewriter";
})(FontStyle || (FontStyle = {}));
class Formula extends Token {
    constructor(parent, prefix, suffix, style) {
        super();
        this.tokens = [];
        this.prefix = "";
        this.suffix = "";
        this.style = FontStyle.Normal;
        this.prev = (t) => {
            var i = this.tokens.indexOf(t);
            if (i < 0 || i == 0)
                return null;
            return this.tokens[i - 1];
        };
        this.next = (t) => {
            var i = this.tokens.indexOf(t);
            if (i < 0 || i == this.tokens.length - 1)
                return null;
            return this.tokens[i + 1];
        };
        this.indexOf = (t) => {
            return this.tokens.indexOf(t);
        };
        this.empty = () => {
            return this.tokens.length == 0;
        };
        this.parent = parent;
        if (prefix !== undefined)
            this.prefix = prefix;
        if (suffix !== undefined)
            this.suffix = suffix;
        if (style !== undefined)
            this.style = style;
    }
    token(i) {
        if (i < 0 || i >= this.tokens.length)
            console.error("[Formula.token] out of range : " + i);
        return this.tokens[i];
    }
    count() {
        return this.tokens.length;
    }
    insert(i, t) {
        this.tokens.splice(i, 0, t);
    }
    remove(from, to) {
        var i, c;
        if (to === undefined)
            i = from, c = 1;
        else {
            i = Math.min(from, to);
            c = Math.abs(to - from);
        }
        return this.tokens.splice(i, c);
    }
    copy(from, to) {
        return this.tokens.slice(Math.min(from, to), Math.max(from, to)).map(s => s.clone(null));
    }
    paste(index, tokens) {
        this.tokens = this.tokens.slice(0, index).concat(tokens.map(s => s.clone(this)).concat(this.tokens.slice(index)));
        return index + tokens.length;
    }
    clone(parent) {
        var s = new Formula(parent);
        s.prefix = this.prefix;
        s.suffix = this.suffix;
        s.style = this.style;
        this.tokens.forEach((t, i) => {
            s.tokens[i] = t.clone(s);
        });
        return s;
    }
    toString() {
        return "Formula( " + this.prefix + this.tokens.map(t => t ? t.toString() : "??").join(" ") + this.suffix + " )";
    }
}

function groupBy(a, f) {
    var r = [];
    var b = a.map(x => ({ key: x, value: f(x) }));
    while (b.length > 0) {
        var p = b.shift();
        var s = [p.key];
        for (var j = 0; j < b.length; j++)
            if (b[j].value == p.value)
                s.push(b.splice(j--, 1)[0].key);
        r.push(s);
    }
    return r;
}
function concat(a) {
    return a.reduce((ac, x) => ac.concat(x), []);
}
function repeat(elem, count) {
    return num(count).map(() => elem);
}
function num(count) {
    return range(0, count);
}
function range(from, count, step) {
    var r = [];
    if (step === undefined)
        step = 1;
    for (var i = 0; i < count; i++)
        r.push(from + i * step);
    return r;
}
function normSquared(x, y) {
    return x * x + y * y;
}

var StrokeStyle;
(function (StrokeStyle) {
    StrokeStyle[StrokeStyle["None"] = 0] = "None";
    StrokeStyle[StrokeStyle["Plain"] = 1] = "Plain";
    StrokeStyle[StrokeStyle["Dotted"] = 2] = "Dotted";
    StrokeStyle[StrokeStyle["Dashed"] = 3] = "Dashed";
    StrokeStyle[StrokeStyle["Wavy"] = 4] = "Wavy";
})(StrokeStyle || (StrokeStyle = {}));
var LabelPosotion;
(function (LabelPosotion) {
    LabelPosotion[LabelPosotion["Left"] = 0] = "Left";
    LabelPosotion[LabelPosotion["Middle"] = 1] = "Middle";
    LabelPosotion[LabelPosotion["Right"] = 2] = "Right";
})(LabelPosotion || (LabelPosotion = {}));
class Diagram extends Matrix {
    constructor(parent, rows, cols) {
        super(parent, rows, cols);
        this.type = StructType.Diagram;
        this.arrows = [];
        this.decorations = [];
        for (var i = 0; i < rows; i++) {
            this.arrows.push(num(cols).map(() => []));
            this.decorations.push([]);
        }
    }
    toggleFrame(index) {
        var p = this.pos(index);
        if (this.decorations[p.row][p.col])
            this.decorations[p.row][p.col] = null;
        else
            this.decorations[p.row][p.col] = {
                size: 0, circle: false, double: false, style: StrokeStyle.Plain
            };
    }
    alterFrameStyle(index, toggleCircle, toggleDouble, style) {
        var p = this.pos(index);
        if (!this.decorations[p.row][p.col])
            this.toggleFrame(index);
        if (toggleCircle)
            this.decorations[p.row][p.col].circle = !this.decorations[p.row][p.col].circle;
        if (toggleDouble)
            this.decorations[p.row][p.col].double = !this.decorations[p.row][p.col].double;
        if (style !== undefined)
            this.decorations[p.row][p.col].style = style;
    }
    changeFrameSize(index, increase) {
        var p = this.pos(index);
        if (!this.decorations[p.row][p.col])
            return;
        this.decorations[p.row][p.col].size += (increase ? 1 : -1);
    }
    addArrow(from, to, num, style, head) {
        var p = this.pos(from);
        var a = {
            from: p,
            to: this.pos(to),
            num: num,
            style: style,
            head: head,
            label: new Formula(this),
            labelPos: LabelPosotion.Left
        };
        var i = this.arrows[p.row][p.col].push(a);
        return i - 1;
    }
    removeArrow(from, to, n) {
        var i = this.findArrow(from, to, n);
        return i ? this.arrows[i.row][i.col].splice(i.i, 1)[0] : null;
    }
    labelArrow(from, to, n, pos) {
        var i = this.findArrow(from, to, n);
        if (!i)
            return null;
        this.arrows[i.row][i.col][i.i].labelPos = pos;
        return this.arrows[i.row][i.col][i.i];
    }
    findArrow(from, to, n) {
        var p = this.pos(from);
        var q = this.pos(to);
        var as = this.arrows[p.row][p.col];
        for (var i = 0; i < as.length; i++) {
            if (as[i].to.row == q.row && as[i].to.col == q.col && n-- == 0)
                return { row: p.row, col: p.col, i: i };
        }
        return null;
    }
    countArrow(from, to) {
        var p = this.pos(from);
        var q = this.pos(to);
        var as = this.arrows[p.row][p.col];
        var n = 0;
        for (var i = 0; i < as.length; i++) {
            if (as[i].to.row == q.row && as[i].to.col == q.col)
                n++;
        }
        return n;
    }
    allArrows() {
        return concat(concat(this.arrows));
    }
    remove(from, to, extensive) {
        var r = this.getRectIndex(from, to);
        return [this.cloneRect(r.i1, r.j1, r.i2, r.j2, true, extensive)];
    }
    cloneRect(i1, j1, i2, j2, erase, extensive) {
        var d = new Diagram(null, i2 - i1 + 1, j2 - j1 + 1);
        for (var i = 0; i < d.rows; i++)
            for (var j = 0; j < d.cols; j++) {
                d.tokenAt(i, j, this.tokenAt(i + i1, j + j1).clone(d));
                if (erase)
                    this.tokenAt(i + i1, j + j1, new Formula(this));
                if (this.decorations[i1 + i][j1 + j]) {
                    d.decorations[i][j] = this.decorations[i1 + i][j1 + j];
                    if (erase)
                        this.decorations[i1 + i][j1 + j] = null;
                }
            }
        for (var i = 0; i < this.arrows.length; i++)
            for (var j = 0; j < this.arrows[i].length; j++)
                for (var k = 0; k < this.arrows[i][j].length; k++) {
                    var a = this.arrows[i][j][k];
                    var start = a.from.row >= i1 && a.from.row <= i2 && a.from.col >= j1 && a.from.col <= j2;
                    var end = a.to.row >= i1 && a.to.row <= i2 && a.to.col >= j1 && a.to.col <= j2;
                    if (!extensive && start && end || extensive && (start || end)) {
                        var b = {
                            from: { row: a.from.row - i1, col: a.from.col - j1 },
                            to: { row: a.to.row - i1, col: a.to.col - j1 },
                            num: a.num, style: a.style, head: a.head, label: a.label.clone(d), labelPos: a.labelPos
                        };
                        if (!extensive || start)
                            d.arrows[b.from.row][b.from.col].push(b);
                        else {
                            b.num *= -1;
                            d.arrows[b.to.row][b.to.col].push(b);
                        }
                        if (erase) {
                            this.arrows[i][j].splice(k, 1);
                            k--;
                        }
                    }
                }
        return d;
    }
    drawFrame(ctx, box, index, deco, color) {
        var a = this.elems[index].renderedElem[0];
        var rect = a.getBoundingClientRect();
        ctx.save();
        ctx.translate(-box.left, -box.top);
        Diagram.setStyle(ctx, deco.style, color);
        if (color)
            ctx.strokeStyle = color;
        if (deco.circle) {
            ctx.beginPath();
            var x = (rect.left + rect.right) / 2, y = (rect.top + rect.bottom) / 2;
            var r = Math.max(rect.width, rect.height) / 2 + 3 * (deco.size - 2);
            ctx.arc(x, y, r, 0, 2 * Math.PI);
            if (deco.double)
                ctx.arc(x, y, r - 3, 0, 2 * Math.PI);
            ctx.stroke();
        }
        else {
            var x = rect.left - 3 * deco.size;
            var y = rect.top - 3 * deco.size;
            var w = rect.width + 6 * deco.size;
            var h = rect.height + 6 * deco.size;
            ctx.strokeRect(x, y, w, h);
            if (deco.double)
                ctx.strokeRect(x + 3, y + 3, w - 6, h - 6);
        }
        ctx.restore();
    }
    drawArrow(ctx, box, label, arrow, shift, color) {
        var len = (x, y) => Math.sqrt(x * x + y * y);
        var a = this.tokenAt(arrow.from.row, arrow.from.col).renderedElem[0];
        var b = this.tokenAt(arrow.to.row, arrow.to.col).renderedElem[0];
        var rec1 = a.getBoundingClientRect();
        var rec2 = b.getBoundingClientRect();
        var acx = (rec1.left + rec1.right) / 2;
        var acy = (rec1.top + rec1.bottom) / 2;
        var bcx = (rec2.left + rec2.right) / 2;
        var bcy = (rec2.top + rec2.bottom) / 2;
        var dx = bcx - acx;
        var dy = bcy - acy;
        var rc = len(dx, dy);
        var arg = Math.atan2(dy, dx);
        var dec1 = this.decorations[arrow.from.row][arrow.from.col];
        var dec2 = this.decorations[arrow.to.row][arrow.to.col];
        var ax, ay;
        var bx, by;
        if (dec1 && dec1.circle) {
            var r1 = Math.max(rec1.width, rec1.height) / 2 + 3 * (dec1.size - 2);
            ax = acx + r1 * dx / rc;
            ay = acy + r1 * dy / rc;
        }
        else if (dy / dx < rec1.height / rec1.width && dy / dx > -rec1.height / rec1.width) {
            var sgn = (dx > 0 ? 1 : -1);
            var w = rec1.width / 2 + 3 * (dec1 ? dec1.size : 0);
            ax = acx + sgn * w;
            ay = acy + sgn * dy / dx * w;
        }
        else {
            var sgn = (dy > 0 ? 1 : -1);
            var h = rec1.height / 2 + 3 * (dec1 ? dec1.size : 0);
            ax = acx + sgn * dx / dy * h;
            ay = acy + sgn * h;
        }
        var r = len(bcx - ax, bcy - ay);
        if (dec2 && dec2.circle)
            r -= Math.max(rec2.width, rec2.height) / 2 + 3 * (dec2.size - 2);
        else if (dy / dx < rec2.height / rec2.width && dy / dx > -rec2.height / rec2.width)
            r -= (rec2.width / 2 + 3 * (dec2 ? dec2.size : 0)) * Math.sqrt(1 + dy * dy / (dx * dx));
        else
            r -= (rec2.height / 2 + 3 * (dec2 ? dec2.size : 0)) * Math.sqrt(1 + dx * dx / (dy * dy));
        ctx.save();
        if (color)
            ctx.strokeStyle = color;
        ctx.beginPath();
        ctx.translate(ax - box.left, ay - box.top);
        var adj = 3; // for wavy arrow (pattern adjustment)
        ctx.rotate(arg);
        ctx.translate(0, -adj + shift);
        ctx.save();
        Diagram.setStyle(ctx, arrow.style, color);
        var headOpen = 1 + 3 * arrow.num;
        var y = adj - Diagram.interShaft * (arrow.num - 1) / 2;
        for (var i = 0; i < arrow.num; i++) {
            ctx.moveTo(0, y);
            ctx.lineTo(r - 14 / headOpen * Diagram.interShaft * Math.abs(i - (arrow.num - 1) / 2), y);
            y += Diagram.interShaft;
        }
        ctx.stroke();
        ctx.restore();
        switch (arrow.head) {
            case ">":
                ctx.beginPath();
                ctx.moveTo(r, adj);
                ctx.bezierCurveTo(r - 5, adj, r - 8, adj - headOpen / 2, r - 10, adj - headOpen);
                ctx.moveTo(r, adj);
                ctx.bezierCurveTo(r - 5, adj, r - 8, adj + headOpen / 2, r - 10, adj + headOpen);
                ctx.stroke();
                break;
        }
        ctx.restore();
        if (arrow.label !== null && label !== null) {
            var lrec = arrow.label.renderedElem[0].getBoundingClientRect();
            var ldiag = Math.sqrt(lrec.width * lrec.width + lrec.height * lrec.height) / 2;
            var x = (acx + bcx) / 2 - lrec.width / 2 - box.left;
            var y = (acy + bcy) / 2 - lrec.height / 2 - box.top;
            if (arrow.labelPos != LabelPosotion.Middle) {
                var t = Math.atan2(-dy, dx)
                    + (arrow.labelPos == LabelPosotion.Left ? 1 : -1) * Math.PI / 2;
                x += (lrec.width + 2) / 2 * Math.cos(t);
                y -= lrec.height / 2 * Math.sin(t);
            }
            else {
                ctx.fillStyle = "#fff";
                ctx.fillRect(x, y, lrec.width + 2, lrec.height);
            }
            label.css({
                "left": x,
                "top": y
            });
        }
    }
    static setStyle(ctx, style, color) {
        switch (style) {
            case StrokeStyle.None:
                ctx.globalAlpha = 0.2;
                break;
            case StrokeStyle.Dashed:
                Diagram.setLineDash(ctx, [5, 5]);
                break;
            case StrokeStyle.Dotted:
                Diagram.setLineDash(ctx, [2, 2]);
                break;
            case StrokeStyle.Wavy:
                ctx.strokeStyle = Diagram.wavyPattern(ctx, color);
                ctx.lineWidth = 6;
                break;
        }
    }
    static wavyPattern(ctx, color) {
        if (color === undefined)
            color = "#000";
        if (!(color in Diagram.wavy)) {
            var canv = document.createElement("canvas");
            canv.width = 12;
            canv.height = 6;
            var c = canv.getContext("2d");
            c.strokeStyle = color;
            c.beginPath();
            c.moveTo(0, 3);
            c.bezierCurveTo(4, -2, 8, 8, 12, 3);
            c.stroke();
            Diagram.wavy[color] = canv;
        }
        return ctx.createPattern(Diagram.wavy[color], "repeat-x");
    }
    static setLineDash(ctx, a) {
        if (ctx.setLineDash !== undefined)
            ctx.setLineDash(a);
        else if (ctx["mozDash"] !== undefined)
            ctx["mozDash"] = a;
    }
    extend(horizontal) {
        if (horizontal) {
            this.arrows.forEach(r => r.push([]));
        }
        else {
            this.arrows.push(num(this.cols).map(() => []));
            this.decorations.push([]);
        }
        super.extend(horizontal);
    }
    shrink(horizontal) {
        if (horizontal) {
            this.arrows.forEach(r => r.splice(this.cols - 1, 1));
            this.decorations.forEach(r => r.splice(this.cols - 1, 1));
        }
        if (!horizontal) {
            this.arrows.pop();
            this.decorations.pop();
        }
        super.shrink(horizontal);
    }
    paste(index, tokens) {
        var n = super.paste(index, tokens);
        if (tokens.length != 1 || !(tokens[0] instanceof Diagram))
            return n;
        var d = tokens[0];
        var p = this.pos(index);
        var r = Math.min(d.rows, this.rows - p.row);
        var c = Math.min(d.cols, this.cols - p.col);
        for (var i = 0; i < r; i++)
            for (var j = 0; j < c; j++)
                this.decorations[p.row + i][p.col + j] = d.decorations[i][j];
        d.arrows.forEach((ar, i) => ar.forEach((ac, j) => ac.forEach(a => {
            var b = {
                from: { row: a.from.row + p.row, col: a.from.col + p.col },
                to: { row: a.to.row + p.row, col: a.to.col + p.col },
                num: Math.abs(a.num), style: a.style, head: a.head, label: a.label.clone(this), labelPos: a.labelPos
            };
            if (a.num < 0)
                this.arrows[b.to.row][b.to.col].push(b);
            else
                this.arrows[b.from.row][b.from.col].push(b);
        })));
        return n;
    }
    clone(parent) {
        var d = this.cloneRect(0, 0, this.rows - 1, this.cols - 1, false, false);
        d.parent = parent;
        return d;
    }
    nonEmpty(i0, j0, rows, cols) {
        return super.nonEmpty(i0, j0, rows, cols)
            || this.decorations.some((r, i) => r.some((d, j) => i >= i0 && i - i0 < rows && j >= j0 && j - j0 < cols))
            || this.allArrows().some(a => a.from.row >= i0 && a.from.row - i0 < rows && a.from.col >= j0 && a.from.col - j0 < cols
                || a.to.row >= i0 && a.to.row - i0 < rows && a.to.col >= j0 && a.to.col - j0 < cols);
    }
    toString() {
        return "Diagram" + this.rows + "," + this.cols
            + "(a:" + this.arrows.reduce((s, r) => s + r.reduce((t, c) => t + c.length, 0), 0)
            + ",d:" + this.decorations.reduce((s, r) => s + r.filter(d => d ? true : false).length, 0)
            + ")[" + this.elems.map(f => f.toString()).join(", ") + "]";
    }
}
Diagram.wavy = {};
Diagram.interShaft = 3;

const symbols = {
    "̀": "grave",
    "́": "acute",
    "̂": "hat",
    "̃": "tilde",
    "̄": "bar",
    "̆": "breve",
    "̇": "dot",
    "̈": "ddot",
    "̊": "mathring",
    "̌": "check",
    "arccos": "arccos",
    "arcsin": "arcsin",
    "arctan": "arctan",
    "arg": "arg",
    "cos": "cos",
    "cosh": "cosh",
    "cot": "cot",
    "coth": "coth",
    "csc": "csc",
    "deg": "deg",
    "det": "det",
    "dim": "dim",
    "exp": "exp",
    "gcd": "gcd",
    "hom": "hom",
    "inf": "inf",
    "ker": "ker",
    "lg": "lg",
    "lim": "lim",
    "liminf": "liminf",
    "limsup": "limsup",
    "ln": "ln",
    "log": "log",
    "max": "max",
    "min": "min",
    "Pr": "Pr",
    "sec": "sec",
    "sin": "sin",
    "sinh": "sinh",
    "sup": "sup",
    "tan": "tan",
    "tanh": "tanh",
    "mod": "bmod",
    "‖": "Vert",
    "{": "{",
    "}": "}",
    "ł": "l",
    "Ł": "L",
    "ø": "o",
    "Ø": "O",
    "ı": "i",
    "ȷ": "j",
    "ß": "ss",
    "æ": "ae",
    "Æ": "AE",
    "œ": "oe",
    "Œ": "OE",
    "å": "aa",
    "Å": "AA",
    "©": "copyright",
    "£": "pounds",
    "…": "dots",
    "⋯": "cdots",
    "⋮": "vdots",
    "⋱": "ddots",
    "ℓ": "ell",
    "𝚤": "imath",
    "𝚥": "jmath",
    "℘": "wp",
    "ℜ": "Re",
    "ℑ": "Im",
    "ℵ": "aleph",
    "∂": "partial",
    "∞": "infty",
    "′": "prime",
    "∅": "emptyset",
    "\\": "backslash",
    "∀": "forall",
    "∃": "exists",
    //"∫": "smallint",
    "△": "triangle",
    "√": "surd",
    "⟙": "top",
    "⟘": "bot",
    "¶": "P",
    "§": "S",
    "♭": "flat",
    "♮": "natural",
    "♯": "sharp",
    "☘": "clubsuit",
    "♢": "diamondsuit",
    "♡": "heartsuit",
    "♠": "spadesuit",
    "¬": "neg",
    "∇": "nabla",
    "□": "Box",
    "◇": "Diamond",
    "±": "pm",
    "∓": "mp",
    "×": "times",
    "÷": "div",
    "∗": "ast",
    "⋆": "star",
    "◦": "circ",
    "∙": "bullet",
    "·": "cdot",
    "∩": "cap",
    "∪": "cup",
    "⊓": "sqcap",
    "⊔": "sqcup",
    "∨": "vee",
    "∧": "wedge",
    "∖": "setminus",
    "≀": "wr",
    "⋄": "diamond",
    "†": "dagger",
    "‡": "ddagger",
    "⨿": "amalg",
    //"△": "bigtriangleup",
    //"▽": "bigtriangledown",
    "◁": "triangleleft",
    "▷": "triangleright",
    "⊲": "lhd",
    "⊳": "rhd",
    "⊴": "unlhd",
    "⊵": "unrhd",
    "⊎": "uplus",
    "⊕": "oplus",
    "⊖": "ominus",
    "⊗": "otimes",
    "⊘": "oslash",
    "⊙": "odot",
    "○": "bigcirc",
    "∑": "sum",
    "⋂": "bigcap",
    "⨀": "bigodot",
    "∏": "prod",
    "⋃": "bigcup",
    "⨂": "bigotimes",
    "∐": "coprod",
    "⨆": "bigsqcup",
    "⨁": "bigoplus",
    "∫": "int",
    "⋁": "bigvee",
    "⨄": "biguplus",
    "∮": "oint",
    "⋀": "bigwedge",
    "≤": "leq",
    "≥": "geq",
    "≺": "prec",
    "≻": "succ",
    "⪯": "preceq",
    "⪰": "succeq",
    "≪": "ll",
    "≫": "gg",
    "⊂": "subset",
    "⊃": "supset",
    "⊆": "subseteq",
    "⊇": "supseteq",
    "⊑": "sqsubseteq",
    "⊒": "sqsupseteq",
    "∈": "in",
    "∋": "ni",
    "⊢": "vdash",
    "⊣": "dashv",
    "≡": "equiv",
    "⊧": "models",
    "∼": "sim",
    "⟂": "perp",
    "≃": "simeq",
    "∝": "propto",
    "∣": "mid",
    "≍": "asymp",
    "∥": "parallel",
    "≈": "approx",
    "⋈": "bowtie",
    "≅": "cong",
    "⨝": "Join",
    "∉": "notin",
    "≠": "neq",
    "⌣": "smile",
    "≐": "doteq",
    "⌢": "frown",
    "⌊": "lfloor",
    "⌋": "rfloor",
    "⌈": "lceil",
    "⌉": "rceil",
    "〈": "langle",
    "〉": "rangle",
    "←": "leftarrow",
    "⟵": "longleftarrow",
    "⇐": "Leftarrow",
    "⟸": "Longleftarrow",
    "→": "rightarrow",
    "⟶": "longrightarrow",
    "⇒": "Rightarrow",
    "⟹": "Longrightarrow",
    "↔": "leftrightarrow",
    "⟷": "longleftrightarrow",
    "⇔": "Leftrightarrow",
    "⟺": "Longleftrightarrow",
    "↦": "mapsto",
    "⟼": "longmapsto",
    "↩": "hookleftarrow",
    "↪": "hookrightarrow",
    "↼": "leftharpoonup",
    "⇀": "rightharpoonup",
    "↽": "leftharpoondown",
    "⇁": "rightharpoondown",
    "↝": "leadsto",
    "↑": "uparrow",
    "⇑": "Uparrow",
    "↓": "downarrow",
    "⇓": "Downarrow",
    "↕": "updownarrow",
    "⇕": "Updownarrow",
    "↗": "nearrow",
    "↘": "searrow",
    "↙": "swarrow",
    "↖": "nwarrow",
    "α": "alpha",
    "β": "beta",
    "γ": "gamma",
    "δ": "delta",
    "ϵ": "epsilon",
    "ζ": "zeta",
    "η": "eta",
    "θ": "theta",
    "ι": "iota",
    "κ": "kappa",
    "λ": "lambda",
    "μ": "mu",
    "ν": "nu",
    "ξ": "xi",
    "π": "pi",
    "ρ": "rho",
    "σ": "sigma",
    "τ": "tau",
    "υ": "upsilon",
    "ϕ": "phi",
    "χ": "chi",
    "ψ": "psi",
    "ω": "omega",
    "Γ": "Gamma",
    "Δ": "Delta",
    "Θ": "Theta",
    "Λ": "Lambda",
    "Ξ": "Xi",
    "Π": "Pi",
    "Σ": "Sigma",
    "Υ": "Upsilon",
    "Φ": "Phi",
    "Ψ": "Psi",
    "Ω": "Omega",
    "ε": "varepsilon",
    "ϑ": "vartheta",
    "ϖ": "varpi",
    "ϱ": "varrho",
    "ς": "varsigma",
    "φ": "varphi",
    "ħ": "hbar",
    "ℏ": "hslash",
    "𝕜": "Bbbk",
    //"□": "square",
    "■": "blacksquare",
    "Ⓢ": "circledS",
    "▲": "blacktriangle",
    "▽": "triangledown",
    "▼": "blacktriangledown",
    "∁": "complement",
    "⅁": "Game",
    "◊": "lozenge",
    "⧫": "blacklozenge",
    "★": "bigstar",
    "∠": "angle",
    "∡": "measuredangle",
    "∢": "sphericalangle",
    "／": "diagup",
    "＼": "diagdown",
    "‵": "backprime",
    "Ⅎ": "Finv",
    "℧": "mho",
    "ð": "eth",
    "∄": "nexists",
    "≑": "doteqdot",
    "≓": "risingdotseq",
    "≒": "fallingdotseq",
    "≖": "eqcirc",
    "≗": "circeq",
    "≏": "bumpeq",
    "≎": "Bumpeq",
    "⋖": "lessdot",
    "⋗": "gtrdot",
    "⩽": "leqslang",
    "⩾": "geqslant",
    "⪕": "eqslantless",
    "⪖": "eqslantgtr",
    "≦": "leqq",
    "≧": "geqq",
    "⋘": "lll",
    "⋙": "ggg",
    "≲": "lesssim",
    "≳": "gtrsim",
    "⪅": "lessapprox",
    "⪆": "gtrapprox",
    "≶": "lessgtr",
    "≷": "gtrless",
    "⋚": "lesseqgtr",
    "⋛": "gtreqless",
    "⪋": "lesseqqgtr",
    "⪌": "gtreqqless",
    "∽": "backsim",
    "⋍": "backsimeq",
    "≼": "preccurlyeq",
    "≽": "succcurlyeq",
    "≊": "approxeq",
    "⋞": "curlyeqprec",
    "⋟": "curlyeqsucc",
    "≾": "precsim",
    "≿": "succsim",
    "⪷": "precapprox",
    "⪸": "succapprox",
    "⫅": "subseteqq",
    "⫆": "supseteqq",
    "⋐": "Subset",
    "⋑": "Supset",
    "⊏": "sqsubset",
    "⊐": "sqsupset",
    "⊨": "vDash",
    "⊩": "Vdash",
    "⊪": "Vvdash",
    "϶": "backepsilon",
    "∴": "therefore",
    "∵": "because",
    "≬": "between",
    "⋔": "pitchfork",
    //"⊲": "vartriangleleft",
    //"⊳": "vartriangleright",
    "◀": "blacktriangleleft",
    "▶": "blacktriangleright",
    //"⊴": "trianglelefteq",
    //"⊵": "trianglerighteq",
    "∔": "dotplus",
    //"·": "centerdot",
    "⋉": "ltimes",
    "⋊": "rtimes",
    "⋋": "leftthreetimes",
    "⋌": "rightthreetimes",
    "⊝": "circleddash",
    "⌅": "barwedge",
    "⌆": "doubebarwedge",
    "⋏": "curlywedge",
    "⋎": "curlyvee",
    "⊻": "veebar",
    "⊺": "intercal",
    "⋒": "Cap",
    "⋓": "Cup",
    "⊛": "circledast",
    "⊚": "circledcirc",
    "⊟": "boxminus",
    "⊠": "boxtimes",
    "⊡": "boxdot",
    "⊞": "boxplus",
    "⋇": "divideontimes",
    "&": "And",
    "∬": "iint",
    "∭": "iiint",
    "⨌": "iiiint",
    "⤎": "dashleftarrow",
    "⤏": "dashrightarrow",
    "⇇": "leftleftarrows",
    "⇉": "rightrightarrows",
    "⇈": "upuparrows",
    "⇊": "downdownarrows",
    "⇆": "leftrightarrows",
    "⇄": "rightleftarrows",
    "⇚": "Lleftarrow",
    "⇛": "Rrightarrow",
    "↿": "upharpoonleft",
    "↾": "upharpoonright",
    "↞": "twoheadleftarrow",
    "↠": "twoheadrightarrow",
    "↢": "leftarrowtail",
    "↣": "rightarrowtail",
    "⇃": "downharpoonleft",
    "⇂": "downharpoonright",
    "⇋": "leftrightharpoons",
    "⇌": "rightleftharpoons",
    "↰": "Lsh",
    "↱": "Rsh",
    "⇝": "rightsquigarrow",
    "↫": "looparrowleft",
    "↬": "looparrowright",
    "↭": "leftrightsquigarrow",
    "⊸": "multimap",
    "↶": "curvearrowleft",
    "↷": "curvearrowright",
    "↺": "circlearrowleft",
    "↻": "circlearrowright",
    "≮": "nless",
    "≯": "ngtr",
    "⪇": "lneq",
    "⪈": "gneq",
    "≰": "nleq",
    "≱": "ngeq",
    "≨": "lneqq",
    "≩": "gneqq",
    "∤": "nmid",
    "∦": "nparallel",
    "⋦": "lnsim",
    "⋧": "gnsim",
    "≁": "nsim",
    "≇": "ncong",
    "⊀": "nprec",
    "⊁": "nsucc",
    "⪵": "precneqq",
    "⪶": "succneqq",
    "⋨": "precnsim",
    "⋩": "succnsim",
    "⪉": "precnapprox",
    "⪊": "succnapprox",
    "⊊": "subsetneq",
    "⊋": "supsetneq",
    "⋪": "ntriangleleft",
    "⋫": "ntriangleright",
    "⋬": "ntrianglelefteq",
    "⋭": "ntrianglerighteq",
    "⊈": "nsubseteq",
    "⊉": "nsupseteq",
    "⫋": "subsetneqq",
    "⫌": "supsetneqq",
    "⊬": "nvdash",
    "⊭": "nvDash",
    "⊮": "nVdash",
    "⊯": "nVDash",
    "↚": "nleftarrow",
    "↛": "nrightarrow",
    "↮": "nleftrightarrow",
    "⇎": "nLeftrightarrow",
    "⇍": "nLeftarrow",
    "⇏": "nRightarrow"
};

let proofMode = false;
const amsBracket = {
    "()": "p",
    "[]": "b",
    "{}": "B",
    "||": "v",
    "‖‖": "V",
    "": ""
};
const styles = {
    "mathbf": FontStyle.Bold,
    "mathrm": FontStyle.Roman,
    "mathscr": FontStyle.Script,
    "mathfrak": FontStyle.Fraktur,
    "mathbb": FontStyle.BlackBoard,
    "mathtt": FontStyle.Typewriter
};
const accentSymbols = {
    "←": "overleftarrow",
    "→": "overrightarrow",
    "～": "widetilde",
    "＾": "widehat",
    "‾": "overline",
    "_": "underline",
    "︷": "overbrace",
    "︸": "underbrace"
};
let combiningAccents = ["̀", "́", "̂", "̃", "̄", "̆", "̇", "̈", "̊", "̌"];
function macro(n, ...args) {
    return "\\" + n + "{ " + args.map(t => trans(t)).join(" }{ ") + " }";
}
function macroBroken(n, indent, ...args) {
    var inner = indent + "  ";
    return "\\" + n + " {\n"
        + inner + args.map(t => trans(t, inner)).join("\n" + indent + "}{\n" + inner) + "\n"
        + indent + "}";
}
function trans(t, indent, proof) {
    if (proof != undefined)
        proofMode = proof;
    if (indent == undefined)
        indent = "";
    if (t instanceof Symbol) {
        return transSymbol(t.str, indent);
    }
    else if (t instanceof Num) {
        return t.value.toString();
    }
    else if (t instanceof Macro) {
        var mc = t;
        return "\\" + mc.name
            + (mc.elems.length > 0
                ? "{ " + mc.elems.map(e => trans(e)).join(" }{ ") + " }" : "");
    }
    else if (t instanceof Diagram) {
        var d = t;
        return "\\xymatrix {"
            + transDiagram(d, indent)
            + "}";
    }
    else if (t instanceof Matrix) {
        var m = t;
        var opt = "";
        for (var i = 0; i < m.cols; i++)
            opt += "c";
        return "\\begin{array}{" + opt + "}"
            + transMatrix(m, indent)
            + "\\end{array}";
    }
    else if (t instanceof Structure) {
        return transStructure(t, indent);
    }
    else if (t instanceof Formula) {
        return transFormula(t, indent);
    }
    else
        return "?";
}
function transSymbol(str, indent) {
    var s = str.charAt(str.length - 1);
    if (combiningAccents.indexOf(s) >= 0) {
        return "\\" + symbols[s] + "{" + transSymbol(str.slice(0, -1), indent) + "}";
    }
    if (proofMode) {
        switch (str) {
            case "&":
                return "&\n" + indent.slice(0, -1);
            case "∧":
                return "\\land";
            case "∨":
                return "\\lor";
            case "¬":
            case "￢":
                return "\\lnot";
        }
    }
    if (str in symbols)
        return "\\" + symbols[str];
    else
        return str;
}
function transDiagram(d, indent) {
    var ln = "\n";
    var str = "";
    str += ln;
    for (var i = 0; i < d.rows; i++) {
        str += d.elems.slice(d.cols * i, d.cols * (i + 1))
            .map((o, j) => {
            var s = trans(o);
            var dec = transDecoration(d.decorations[i][j]);
            if (dec != "")
                s = "*" + dec + "{" + s + "}";
            s += groupBy(d.arrows[i][j], a => a.to.row * d.cols + a.to.col)
                .map(as => as.map((a, k) => transArrow(a, k - (as.length - 1) / 2)).join(" ")).join(" ");
            return s;
        }).join(" & ") + " \\\\" + ln;
    }
    return str;
}
function transDecoration(deco) {
    if (!deco)
        return "";
    var s = "";
    if (deco.size != 0)
        s = repeat((deco.size > 0 ? "+" : "-"), Math.abs(deco.size)).join("");
    if (deco.circle)
        s += "[o]";
    switch (deco.style) {
        case StrokeStyle.Plain:
            s += "[F" + (deco.double ? "=" : "-") + "]";
            break;
        case StrokeStyle.Dashed:
            s += "[F--]";
            break;
        case StrokeStyle.Dotted:
            s += "[F.]";
            break;
        case StrokeStyle.Wavy:
            s += "[F~]";
            break;
    }
    return s;
}
function transArrow(a, shift) {
    var s = "";
    var style = "";
    var doubled = false;
    switch (a.style) {
        case StrokeStyle.Plain:
            style = ((doubled = a.num == 2) ? "=" : "-");
            break;
        case StrokeStyle.Dashed:
            style = ((doubled = a.num == 2) ? "==" : "--");
            break;
        case StrokeStyle.Dotted:
            style = ((doubled = a.num == 2) ? ":" : ".");
            break;
        case StrokeStyle.Wavy:
            style = "~";
            break;
    }
    style += a.head;
    if (style != "->")
        style = "{" + style + "}";
    else
        style = "";
    if (!doubled && a.num != 1)
        style = a.num.toString() + style;
    if (shift)
        style += "<" + shift.toString() + "ex>";
    s += (style != "" ? "\\ar@" + style : "\\ar");
    var dir = "";
    var dc = a.to.col - a.from.col;
    var dr = a.to.row - a.from.row;
    if (dc != 0)
        dir = repeat((dc > 0 ? "r" : "l"), Math.abs(dc)).join();
    if (dr != 0)
        dir += repeat((dr > 0 ? "d" : "u"), Math.abs(dr)).join();
    s += "[" + dir + "]";
    if (!a.label.empty()) {
        if (a.labelPos == LabelPosotion.Middle)
            s += " |";
        else if (a.labelPos == LabelPosotion.Right)
            s += "_";
        else
            s += "^";
        var t = trans(a.label);
        if (t.length > 1)
            t = "{" + t + "}";
        s += t;
    }
    return s;
}
function transMatrix(m, indent) {
    var ln = (m.rows >= 2 && m.cols >= 2 && !(m.rows == 2 && m.cols == 2))
        ? "\n" : " ";
    var str = "";
    str += ln;
    for (var i = 0; i < m.rows; i++) {
        str += m.elems.slice(m.cols * i, m.cols * (i + 1))
            .map(f => trans(f)).join(" & ")
            + " \\\\" + ln;
    }
    return str;
}
function transStructure(s, indent) {
    var str;
    switch (s.type) {
        case StructType.Frac:
            return macro("frac", s.token(0), s.token(1));
        case StructType.Infer:
            var opt = trans(s.token(2));
            return macroBroken("infer" + (opt != "" ? "[" + opt + "]" : ""), indent, s.token(0), s.token(1));
        case StructType.Power:
            str = trans(s.token(0));
            return str.length == 1
                ? "^" + str
                : "^{ " + str + " }";
        case StructType.Index:
            str = trans(s.token(0));
            return str.length == 1
                ? "_" + str
                : "_{ " + str + " }";
        case StructType.BigOpr:
            return transSymbol(s.operator, indent)
                + "_{" + trans(s.elems[0])
                + "}^{" + trans(s.elems[1]) + "}";
            break;
        case StructType.Accent:
            return "\\" + accentSymbols[s.symbol]
                + "{" + trans(s.elems[0]) + "}";
            break;
        default:
            return "?struct?";
    }
}
function transFormula(f, indent) {
    if (f.tokens.length == 1 && f.tokens[0] instanceof Matrix
        && !(f.tokens[0] instanceof Diagram)) {
        var br = f.prefix + f.suffix;
        if (br in amsBracket) {
            var n = amsBracket[br];
            return "\\begin{" + n + "matrix}"
                + transMatrix(f.tokens[0], indent)
                + "\\end{" + n + "matrix}";
        }
    }
    var separator = " ";
    var pre, suf;
    if (f.style != FontStyle.Normal) {
        var cmd;
        for (cmd in styles)
            if (styles[cmd] == f.style) {
                pre = "\\" + cmd + "{";
                suf = "}";
                break;
            }
        if (f.tokens.every(t => t instanceof Symbol || t instanceof Num))
            separator = "";
    }
    else if (f.prefix == "√" && f.suffix == "") {
        pre = "\\sqrt{ ";
        suf = " }";
    }
    else {
        pre = transSymbol(f.prefix, indent);
        suf = transSymbol(f.suffix, indent);
        if (pre != "")
            pre = "\\left" + pre + " ";
        else if (suf != "")
            pre = "\\left. ";
        if (suf != "")
            suf = " \\right" + suf;
        else if (pre != "")
            suf = " \\right.";
    }
    return pre + f.tokens.map(t => trans(t, indent)).join(separator) + suf;
}

var LaTeXASTType;
(function (LaTeXASTType) {
    LaTeXASTType[LaTeXASTType["Sequence"] = 0] = "Sequence";
    LaTeXASTType[LaTeXASTType["Environment"] = 1] = "Environment";
    LaTeXASTType[LaTeXASTType["Command"] = 2] = "Command";
    LaTeXASTType[LaTeXASTType["Symbol"] = 3] = "Symbol";
    LaTeXASTType[LaTeXASTType["Number"] = 4] = "Number";
})(LaTeXASTType || (LaTeXASTType = {}));
class LaTeXReader {
    constructor(src) {
        this.macroArgNum = {};
        this.rest = src;
    }
    static parse(source) {
        var parser = new LaTeXReader(source);
        return parser.parseSeq();
    }
    parseSeq(eof) {
        var tokens = [];
        console.debug("parseSeq " + this.rest.substr(0, 8) + " ...");
        while (this.rest.length > 0) {
            var t = this.parseToken(eof);
            if (!t)
                break;
            console.debug("parsed " + t.value + " -- " + this.rest.substr(0, 8) + " ...");
            tokens.push(t);
        }
        console.debug("exit seq");
        if (tokens.length == 1)
            return tokens[0];
        else
            return {
                type: LaTeXASTType.Sequence,
                value: "",
                children: tokens
            };
    }
    parseMatrix(isXy) {
        var mat = [];
        var row = [];
        var cell = [];
        var m;
        while (this.rest) {
            this.white();
            if (this.str("\\\\")) {
                row.push({ type: LaTeXASTType.Sequence, value: "", children: cell });
                mat.push({ type: LaTeXASTType.Sequence, value: "", children: row });
                cell = [];
                row = [];
            }
            else if (this.str("&")) {
                row.push({ type: LaTeXASTType.Sequence, value: "", children: cell });
                cell = [];
            }
            else if (isXy && (m = this.pattern(/^\*\s*(\+*|-*)\s*(?:\[(o)\])?\s*(?:\[([^\]]*)\])?/))) {
                console.debug("*");
                var item = this.parseToken();
                cell.push({
                    type: LaTeXASTType.Command,
                    value: "*",
                    children: [
                        this.optionalSymbol(m[1]),
                        this.optionalSymbol(m[2]),
                        this.optionalSymbol(m[3]),
                        item]
                });
            }
            else if (isXy && (m = this.pattern(/^\\ar(?:@([2-9])?(?:\{([^\}]*)\})?)?\[([^\]]*)\]\s*([\^\|_])?/))) {
                var args = [
                    this.optionalSymbol(m[1]),
                    this.optionalSymbol(m[2]),
                    this.optionalSymbol(m[3]),
                    this.optionalSymbol(m[4]),
                ];
                if (m[4])
                    args.push(this.parseToken());
                cell.push({
                    type: LaTeXASTType.Command,
                    value: "ar",
                    children: args
                });
            }
            else {
                var t = this.parseToken();
                if (!t)
                    break;
                cell.push(t);
            }
        }
        if (cell.length > 0)
            row.push({ type: LaTeXASTType.Sequence, value: "", children: cell });
        if (row.length > 0)
            mat.push({ type: LaTeXASTType.Sequence, value: "", children: row });
        return mat;
    }
    optionalSymbol(s) {
        return s || s === ""
            ? { type: LaTeXASTType.Symbol, value: s, children: null }
            : null;
    }
    parseToken(eof) {
        var m;
        this.white();
        if (eof && this.str(eof))
            return;
        console.debug("parseToken " + this.rest.charAt(0));
        if (this.str("\\")) {
            if (this.str("\\")) {
                return {
                    type: LaTeXASTType.Symbol,
                    value: "\\\\",
                    children: null
                };
            }
            else if (m = this.pattern(/^[a-zA-Z]+/)) {
                return this.parseCommand(m[0]);
            }
        }
        else if (this.str("^") || this.str("_")) {
            return {
                type: LaTeXASTType.Command,
                value: this.parsed,
                children: [this.parseToken()]
            };
        }
        else if (this.str("#")) {
            m = this.pattern(/^[0-9]+/);
            return {
                type: LaTeXASTType.Symbol,
                value: "#" + this.parsed,
                children: null
            };
        }
        else if (this.str("{")) {
            return this.parseSeq();
        }
        else if (this.str("}")) {
            return null;
        }
        else if (this.pattern(/^[0-9]/)) {
            return {
                type: LaTeXASTType.Number,
                value: this.parsed,
                children: null
            };
        }
        else {
            return {
                type: LaTeXASTType.Symbol,
                value: this.head(),
                children: null
            };
        }
    }
    parseCommand(name) {
        var m;
        if (name == "begin") {
            m = this.pattern(/\{([a-zA-Z0-9]+\*?)\}/);
            if (!m)
                console.error("[LaTeXReader.parseToken] begin command must have 1 arg.");
            var env = (m[1].indexOf("matrix") >= 0
                ? this.parseMatrix(false)
                : [this.parseSeq()]);
            return {
                type: LaTeXASTType.Environment,
                value: m[1],
                children: env
            };
        }
        else if (name == "end") {
            m = this.pattern(/\{[a-zA-Z0-9]+\*?\}/);
            return null;
        }
        else if (name == "xymatrix") {
            this.white();
            this.str("{");
            return {
                type: LaTeXASTType.Command,
                value: name,
                children: this.parseMatrix(true)
            };
        }
        else {
            var ob = this.getArgObligation(name);
            var arg = [];
            for (var i = 0; i < ob.length; i++) {
                if (ob[i])
                    arg.push(this.parseToken());
                else {
                    if (this.str("["))
                        arg.push(this.parseSeq("]"));
                    else
                        arg.push(null);
                }
            }
            if (name == "newcommand" && arg[1])
                this.macroArgNum[arg[0].value] = parseInt(arg[1].value);
            return {
                type: LaTeXASTType.Command,
                value: name,
                children: arg
            };
        }
    }
    getArgObligation(cmd) {
        switch (cmd) {
            case "newcommand":
                return [true, false, true];
            case "infer":
                return [false, true, true];
            case "frac":
                return [true, true];
            case "sqrt":
                return [false, true];
            case "xymatrix":
            case "left":
            case "right":
            case "mathbf":
            case "mathrm":
            case "mathscr":
            case "mathfrak":
            case "mathbb":
            case "mathtt":
            case "grave":
            case "acute":
            case "hat":
            case "tilde":
            case "bar":
            case "breve":
            case "dot":
            case "ddot":
            case "mathring":
            case "check":
            case "widetilde":
            case "widehat":
            case "overleftarrow":
            case "overrightarrow":
            case "overline":
            case "underline":
            case "overbrace":
            case "underbrace":
                return [true];
            default:
                if (cmd in this.macroArgNum)
                    return repeat(true, this.macroArgNum[cmd]);
                else
                    return [];
        }
    }
    white() {
        var i;
        for (i = 0; i < this.rest.length; i++) {
            var c = this.rest.charAt(i);
            if (!(c == " " || c == "\n" || c == "\r"))
                break;
        }
        this.rest = this.rest.substr(i);
    }
    head() {
        var c = this.rest.charAt(0);
        this.rest = this.rest.substr(1);
        return c;
    }
    str(s) {
        if (this.rest.substr(0, s.length) == s) {
            this.parsed = s;
            this.rest = this.rest.substr(s.length);
            return true;
        }
        else
            return false;
    }
    pattern(reg) {
        var m = this.rest.match(reg);
        if (m) {
            this.parsed = m[0];
            this.rest = this.rest.substr(m[0].length);
            return m;
        }
        else
            return null;
    }
}

let EnSpace = "\u2002";
let EmSpace = "\u2003";
let SixPerEmSpace = "\u2006";
let Bold = { "A": "𝐀", "B": "𝐁", "C": "𝐂", "D": "𝐃", "E": "𝐄", "F": "𝐅", "G": "𝐆", "H": "𝐇", "I": "𝐈", "J": "𝐉", "K": "𝐊", "L": "𝐋", "M": "𝐌", "N": "𝐍", "O": "𝐎", "P": "𝐏", "Q": "𝐐", "R": "𝐑", "S": "𝐒", "T": "𝐓", "U": "𝐔", "V": "𝐕", "W": "𝐖", "X": "𝐗", "Y": "𝐘", "Z": "𝐙", "a": "𝐚", "b": "𝐛", "c": "𝐜", "d": "𝐝", "e": "𝐞", "f": "𝐟", "g": "𝐠", "h": "𝐡", "i": "𝐢", "j": "𝐣", "k": "𝐤", "l": "𝐥", "m": "𝐦", "n": "𝐧", "o": "𝐨", "p": "𝐩", "q": "𝐪", "r": "𝐫", "s": "𝐬", "t": "𝐭", "u": "𝐮", "v": "𝐯", "w": "𝐰", "x": "𝐱", "y": "𝐲", "z": "𝐳" };
let Script = { "A": "𝒜", "B": "ℬ", "C": "𝒞", "D": "𝒟", "E": "ℰ", "F": "ℱ", "G": "𝒢", "H": "ℋ", "I": "ℐ", "J": "𝒥", "K": "𝒦", "L": "ℒ", "M": "ℳ", "N": "𝒩", "O": "𝒪", "P": "𝒫", "Q": "𝒬", "R": "ℛ", "S": "𝒮", "T": "𝒯", "U": "𝒰", "V": "𝒱", "W": "𝒲", "X": "𝒳", "Y": "𝒴", "Z": "𝒵", "a": "𝒶", "b": "𝒷", "c": "𝒸", "d": "𝒹", "e": "ℯ", "f": "𝒻", "g": "ℊ", "h": "𝒽", "i": "𝒾", "j": "𝒿", "k": "𝓀", "l": "𝓁", "m": "𝓂", "n": "𝓃", "o": "ℴ", "p": "𝓅", "q": "𝓆", "r": "𝓇", "s": "𝓈", "t": "𝓉", "u": "𝓊", "v": "𝓋", "w": "𝓌", "x": "𝓍", "y": "𝓎", "z": "𝓏" };
let Fraktur = { "A": "𝔄", "B": "𝔅", "C": "ℭ", "D": "𝔇", "E": "𝔈", "F": "𝔉", "G": "𝔊", "H": "ℌ", "I": "ℑ", "J": "𝔍", "K": "𝔎", "L": "𝔏", "M": "𝔐", "N": "𝔑", "O": "𝔒", "P": "𝔓", "Q": "𝔔", "R": "ℜ", "S": "𝔖", "T": "𝔗", "U": "𝔘", "V": "𝔙", "W": "𝔚", "X": "𝔛", "Y": "𝔜", "Z": "ℨ", "a": "𝔞", "b": "𝔟", "c": "𝔠", "d": "𝔡", "e": "𝔢", "f": "𝔣", "g": "𝔤", "h": "𝔥", "i": "𝔦", "j": "𝔧", "k": "𝔨", "l": "𝔩", "m": "𝔪", "n": "𝔫", "o": "𝔬", "p": "𝔭", "q": "𝔮", "r": "𝔯", "s": "𝔰", "t": "𝔱", "u": "𝔲", "v": "𝔳", "w": "𝔴", "x": "𝔵", "y": "𝔶", "z": "𝔷" };
let DoubleStruck = { "A": "𝔸", "B": "𝔹", "C": "ℂ", "D": "𝔻", "E": "𝔼", "F": "𝔽", "G": "𝔾", "H": "ℍ", "I": "𝕀", "J": "𝕁", "K": "𝕂", "L": "𝕃", "M": "𝕄", "N": "ℕ", "O": "𝕆", "P": "ℙ", "Q": "ℚ", "R": "ℝ", "S": "𝕊", "T": "𝕋", "U": "𝕌", "V": "𝕍", "W": "𝕎", "X": "𝕏", "Y": "𝕐", "Z": "ℤ", "a": "𝕒", "b": "𝕓", "c": "𝕔", "d": "𝕕", "e": "𝕖", "f": "𝕗", "g": "𝕘", "h": "𝕙", "i": "𝕚", "j": "𝕛", "k": "𝕜", "l": "𝕝", "m": "𝕞", "n": "𝕟", "o": "𝕠", "p": "𝕡", "q": "𝕢", "r": "𝕣", "s": "𝕤", "t": "𝕥", "u": "𝕦", "v": "𝕧", "w": "𝕨", "x": "𝕩", "y": "𝕪", "z": "𝕫" };
let SansSerif = { "A": "𝖠", "B": "𝖡", "C": "𝖢", "D": "𝖣", "E": "𝖤", "F": "𝖥", "G": "𝖦", "H": "𝖧", "I": "𝖨", "J": "𝖩", "K": "𝖪", "L": "𝖫", "M": "𝖬", "N": "𝖭", "O": "𝖮", "P": "𝖯", "Q": "𝖰", "R": "𝖱", "S": "𝖲", "T": "𝖳", "U": "𝖴", "V": "𝖵", "W": "𝖶", "X": "𝖷", "Y": "𝖸", "Z": "𝖹", "a": "𝖺", "b": "𝖻", "c": "𝖼", "d": "𝖽", "e": "𝖾", "f": "𝖿", "g": "𝗀", "h": "𝗁", "i": "𝗂", "j": "𝗃", "k": "𝗄", "l": "𝗅", "m": "𝗆", "n": "𝗇", "o": "𝗈", "p": "𝗉", "q": "𝗊", "r": "𝗋", "s": "𝗌", "t": "𝗍", "u": "𝗎", "v": "𝗏", "w": "𝗐", "x": "𝗑", "y": "𝗒", "z": "𝗓" };
let Monospace = { "A": "𝙰", "B": "𝙱", "C": "𝙲", "D": "𝙳", "E": "𝙴", "F": "𝙵", "G": "𝙶", "H": "𝙷", "I": "𝙸", "J": "𝙹", "K": "𝙺", "L": "𝙻", "M": "𝙼", "N": "𝙽", "O": "𝙾", "P": "𝙿", "Q": "𝚀", "R": "𝚁", "S": "𝚂", "T": "𝚃", "U": "𝚄", "V": "𝚅", "W": "𝚆", "X": "𝚇", "Y": "𝚈", "Z": "𝚉", "a": "𝚊", "b": "𝚋", "c": "𝚌", "d": "𝚍", "e": "𝚎", "f": "𝚏", "g": "𝚐", "h": "𝚑", "i": "𝚒", "j": "𝚓", "k": "𝚔", "l": "𝚕", "m": "𝚖", "n": "𝚗", "o": "𝚘", "p": "𝚙", "q": "𝚚", "r": "𝚛", "s": "𝚜", "t": "𝚝", "u": "𝚞", "v": "𝚟", "w": "𝚠", "x": "𝚡", "y": "𝚢", "z": "𝚣" };

class Segment {
    constructor(p0x, p0y, p1x, p1y, w0, w1) {
        this.p0 = { x: p0x, y: p0y };
        this.p1 = { x: p1x, y: p1y };
        this.weight0 = w0;
        this.weight1 = w1;
    }
    draw(ctx) {
        console.error("[Segment.draw] draw method not implemented");
    }
}
class Line extends Segment {
    constructor(p0x, p0y, p1x, p1y, w0, w1) {
        super(p0x, p0y, p1x, p1y, w0, w1 !== undefined ? w1 : w0);
    }
    toString() {
        return "line " + this.p0.toString() + " " + this.p1.toString();
    }
    draw(ctx) {
        var t = Math.atan2(this.p1.y - this.p0.y, this.p1.x - this.p0.x) + Math.PI / 2.0;
        var w0 = this.weight0 / 2.0;
        var w1 = this.weight1 / 2.0;
        var d0 = { x: w0 * Math.cos(t), y: w0 * Math.sin(t) };
        var d1 = { x: w1 * Math.cos(t), y: w1 * Math.sin(t) };
        ctx.moveTo(this.p0.x + d0.x, this.p0.y + d0.y);
        ctx.lineTo(this.p1.x + d1.x, this.p1.y + d1.y);
        ctx.lineTo(this.p1.x - d1.x, this.p1.y - d1.y);
        ctx.lineTo(this.p0.x - d0.x, this.p0.y - d0.y);
        ctx.closePath();
    }
}
class Bezier extends Segment {
    constructor(p0x, p0y, p1x, p1y, c1x, c1y, c2x, c2y, w0, w1, w) {
        super(p0x, p0y, p1x, p1y, w0, w1);
        this.c1 = { x: c1x, y: c1y };
        this.c2 = { x: c2x, y: c2y };
        this.weight = w;
    }
    draw(ctx) {
        var w = this.weight / 2.0;
        var w0 = this.weight0 / 2.0;
        var w1 = this.weight1 / 2.0;
        var t0 = Math.atan2(this.c1.y - this.p0.y, this.c1.x - this.p0.x) + Math.PI / 2.0;
        var t1 = Math.atan2(this.c2.y - this.p1.y, this.c2.x - this.p1.x) - Math.PI / 2.0;
        var tc = Math.atan2(this.p1.y - this.p0.y, this.p1.x - this.p0.x) + Math.PI / 2.0;
        var d0 = { x: w0 * Math.cos(t0), y: w0 * Math.sin(t0) };
        var d1 = { x: w1 * Math.cos(t1), y: w1 * Math.sin(t1) };
        var dc = { x: w * Math.cos(tc), y: w * Math.sin(tc) };
        ctx.moveTo(this.p0.x + d0.x, this.p0.y + d0.y);
        ctx.bezierCurveTo(this.c1.x + dc.x, this.c1.y + dc.y, this.c2.x + dc.x, this.c2.y + dc.y, this.p1.x + d1.x, this.p1.y + d1.y);
        ctx.lineTo(this.p1.x - d1.x, this.p1.y - d1.y);
        ctx.bezierCurveTo(this.c2.x - dc.x, this.c2.y - dc.y, this.c1.x - dc.x, this.c1.y - dc.y, this.p0.x - d0.x, this.p0.y - d0.y);
        ctx.closePath();
    }
    toString() {
        return "bezier " + this.p0.toString() + " " + this.p1.toString() + " ("
            + this.c1.toString() + " " + this.c2.toString() + ")";
    }
}
class Glyph {
    constructor(w, h, ...s) {
        this.width = w;
        this.height = h;
        this.seg = s;
    }
    toString() {
        return this.seg.map(s => s.toString()).join("\n");
    }
    reflect() {
        var r = new Glyph(this.width, this.height);
        var w = r.width;
        for (var i = 0; i < this.seg.length; i++) {
            var s = this.seg[i];
            if (s instanceof Bezier) {
                var b = s;
                r.seg.push(new Bezier(w - b.p0.x, b.p0.y, w - b.p1.x, b.p1.y, w - b.c1.x, b.c1.y, w - b.c2.x, b.c2.y, b.weight0, b.weight1, b.weight));
            }
            else if (s instanceof Line)
                r.seg.push(new Line(w - s.p0.x, s.p0.y, w - s.p1.x, s.p1.y, s.weight0, s.weight1));
        }
        return r;
    }
    turnRight() {
        var r = new Glyph(this.height, this.width);
        for (var i = 0; i < this.seg.length; i++) {
            var s = this.seg[i];
            if (s instanceof Bezier) {
                var b = s;
                r.seg.push(new Bezier(b.p0.y, b.p0.x, b.p1.y, b.p1.x, b.c1.y, b.c1.x, b.c2.y, b.c2.x, b.weight0, b.weight1, b.weight));
            }
            else if (s instanceof Line)
                r.seg.push(new Line(s.p0.y, s.p0.x, s.p1.y, s.p1.x, s.weight0, s.weight1));
        }
        return r;
    }
}
class GlyphFactory {
    constructor() {
        this.data = {};
        this.cache = {};
        this.canvas = document.createElement("canvas");
        this.data["("] = new Glyph(24, 64, new Bezier(20, 4, 20, 60, 4, 2, 4, 62, 0, 0, 3));
        this.data["{"] = new Glyph(24, 64, new Bezier(20, 4, 4, 32, 2, 0, 16, 32, 0, 0, 4), new Bezier(4, 32, 20, 60, 16, 32, 2, 64, 0, 0, 4));
        this.data["["] = new Glyph(24, 64, new Line(16, 4, 8, 4, 1), new Line(8, 4, 8, 60, 2.5), new Line(8, 60, 16, 60, 1));
        this.data["|"] = new Glyph(24, 64, new Line(12, 4, 12, 60, 2));
        this.data["‖"] = new Glyph(24, 64, new Line(9, 4, 9, 60, 2), new Line(16, 4, 16, 60, 2));
        this.data["⌊"] = new Glyph(24, 64, new Line(8, 4, 8, 60, 2.5), new Line(8, 60, 16, 60, 1));
        this.data["⌈"] = new Glyph(24, 64, new Line(16, 4, 8, 4, 1), new Line(8, 4, 8, 60, 2.5));
        this.data["〈"] = new Glyph(24, 64, new Line(16, 4, 8, 32, 2), new Line(8, 32, 16, 60, 2));
        this.data["√"] = new Glyph(32, 64, new Line(7, 48, 9, 40, 2), new Line(9, 40, 15, 56, 3), new Line(15, 56, 31, 0, 2));
        this.data[")"] = this.data["("].reflect();
        this.data["}"] = this.data["{"].reflect();
        this.data["]"] = this.data["["].reflect();
        this.data["⌋"] = this.data["⌊"].reflect();
        this.data["⌉"] = this.data["⌈"].reflect();
        this.data["〉"] = this.data["〈"].reflect();
        this.data["～"] = new Glyph(64, 24, new Bezier(8, 16, 56, 8, 28, -6, 36, 30, 1, 1, 4));
        this.data["＾"] = new Glyph(64, 24, new Line(8, 16, 32, 8, 1, 3), new Line(32, 8, 56, 16, 3, 1));
        this.data["→"] = new Glyph(64, 24, new Line(4, 12, 60, 12, 1, 1), new Bezier(60, 12, 46, 6, 50, 12, 54, 10, 0, 0, 5), new Bezier(60, 12, 46, 18, 50, 12, 54, 14, 0, 0, 5));
        this.data["←"] = this.data["→"].reflect();
        this.data["︷"] = this.data["{"].turnRight();
        this.data["︸"] = this.data["}"].turnRight();
    }
    generate(char) {
        if (char in this.cache)
            return this.cache[char];
        else if (!(char in this.data))
            return "";
        var glyph = this.data[char];
        this.canvas.width = glyph.width;
        this.canvas.height = glyph.height;
        var context = this.canvas.getContext("2d");
        context.fillStyle = "#000";
        for (var i = 0; i < glyph.seg.length; i++) {
            context.beginPath();
            glyph.seg[i].draw(context);
            context.fill();
        }
        var dat = this.canvas.toDataURL("image/png");
        this.cache[char] = dat;
        return dat;
    }
}

class MathEx {
    static cosh(x) {
        var y = Math.exp(x);
        return (y + 1 / y) / 2;
    }
    static sinh(x) {
        var y = Math.exp(x);
        return (y - 1 / y) / 2;
    }
    static tanh(x) {
        var y = Math.exp(2 * x);
        return (y - 1) / (y + 1);
    }
    static coth(x) {
        var y = Math.exp(2 * x);
        return (y + 1) / (y - 1);
    }
    static csc(x) {
        return 1 / Math.cos(x);
    }
    static sec(x) {
        return 1 / Math.sin(x);
    }
    static cot(x) {
        return 1 / Math.tan(x);
    }
    static lg(x) {
        return Math.log(x) * MathEx.Log10Inv;
    }
}
MathEx.Log10Inv = 1 / Math.log(10);
var OperatorType;
(function (OperatorType) {
    OperatorType[OperatorType["Prefix"] = 0] = "Prefix";
    OperatorType[OperatorType["Suffix"] = 1] = "Suffix";
    OperatorType[OperatorType["Infix"] = 2] = "Infix";
})(OperatorType || (OperatorType = {}));
class Numeric extends Token {
    constructor(n, d, approx) {
        super();
        this.approx = false;
        this.num = d < 0 ? -n : n;
        this.den = Math.abs(d);
        if (approx !== undefined)
            this.approx = approx;
    }
    get value() {
        return this.num / this.den;
    }
    static negate(n) {
        return new Numeric(-n.num, n.den, n.approx);
    }
    static add(m, n) {
        return new Numeric(m.num * n.den + n.num * m.den, m.den * n.den, m.approx || n.approx);
    }
    static sub(m, n) {
        return new Numeric(m.num * n.den + -n.num * m.den, m.den * n.den, m.approx || n.approx);
    }
    static mul(m, n) {
        return new Numeric(m.num * n.num, m.den * n.den, m.approx || n.approx);
    }
    static div(m, n) {
        return new Numeric(m.num * n.den, m.den * n.num, m.approx || n.approx);
    }
    static fromReal(n) {
        var s = n.toString();
        var i;
        if ((i = s.indexOf(".")) >= 0) {
            var x = parseFloat(s);
            var d = Math.pow(10, s.length - i);
            return new Numeric(n * d, d, true);
        }
        else
            return new Numeric(n, 1, true);
    }
    toString() {
        return this.approx
            ? (this.num / this.den).toString()
            : this.num.toString() + "/" + this.den.toString();
    }
}
Numeric.Zero = new Numeric(0, 1, false);
Numeric.One = new Numeric(1, 1, false);
class Term {
    constructor(coeff, exponent) {
        this.coeff = coeff;
        this.exponent = exponent;
        this.n = Object.keys(exponent).length;
    }
    porpotionalTo(t) {
        if (t.n != this.n)
            return false;
        for (var x in this.exponent)
            if (!(x in t.exponent && t.exponent[x] == this.exponent[x]))
                return false;
        return true;
    }
    toString() {
        var str = this.coeff.toString();
        for (var x in this.exponent)
            str += x + "^" + this.exponent[x];
        return str;
    }
}
class Polynomial extends Token {
    constructor(term) {
        super();
        this.term = [];
        var f = {};
        this.term = term;
    }
    static fromSymbol(s) {
        var f = {};
        f[s.str] = 1;
        var t = new Term(Numeric.One, f);
        return new Polynomial([t]);
    }
    static fromNumeric(n) {
        var t = new Term(n, {});
        return new Polynomial([t]);
    }
    static negate(p) {
        return new Polynomial(p.term.map(t => new Term(Numeric.negate(t.coeff), t.exponent)));
    }
    static additionImpl(p, q, sub) {
        var t = [];
        for (var i = 0; i < p.term.length; i++)
            t.push(p.term[i]);
        for (var j = 0; j < q.term.length; j++) {
            var a = q.term[j];
            var found = false;
            for (var i = 0; i < t.length; i++) {
                if (a.porpotionalTo(t[i])) {
                    t[i].coeff = sub
                        ? Numeric.sub(t[i].coeff, a.coeff)
                        : Numeric.add(t[i].coeff, a.coeff);
                    found = true;
                    break;
                }
            }
            if (!found)
                t.push(a);
        }
        for (var i = 0; i < t.length; i++)
            if (!t[i].coeff.approx && t[i].coeff.value == 0)
                t.splice(i--, 1);
        return new Polynomial(t);
    }
    static add(p, q) {
        return Polynomial.additionImpl(p, q, false);
    }
    static sub(p, q) {
        return Polynomial.additionImpl(p, q, true);
    }
    static multiplyImpl(p, q, div) {
        var t = [];
        for (var i = 0; i < p.term.length; i++) {
            for (var j = 0; j < q.term.length; j++) {
                var c = Numeric.mul(p.term[i].coeff, q.term[j].coeff);
                var f = {};
                for (var x in p.term[i].exponent)
                    f[x] = p.term[i].exponent[x];
                for (var x in q.term[j].exponent) {
                    if (x in f)
                        f[x] += q.term[j].exponent[x];
                    else
                        f[x] = q.term[j].exponent[x];
                }
                for (var x in f) {
                    if (f[x] == 0)
                        delete f[x];
                }
                t.push(new Term(c, f));
            }
        }
        return new Polynomial(t);
    }
    static mul(p, q) {
        return Polynomial.multiplyImpl(p, q, false);
    }
    static div(p, q) {
        return Polynomial.multiplyImpl(p, q, true);
    }
    toString() {
        return this.term.map(t => t.toString()).join(" + ").replace("+ -", "- ");
    }
}
function evalToken(t) {
    console.debug("eval start : " + t.toString());
    var r = evalSeq(t);
    console.debug("eval result : " + (r ? r.toString() : "no value"));
    if (r)
        return interpret(r);
    else
        return null;
}
function interpret(t) {
    if (t instanceof Polynomial) {
        return fromPolynomial(t);
    }
    else if (t instanceof Numeric) {
        return [fromNumeric(t)];
    }
    else if (t instanceof Matrix) {
        var f = new Formula(null, "(", ")");
        var m = t;
        for (var i = 0; i < m.elems.length; i++) {
            m.elems[i].tokens = m.elems[i].tokens.reduce((prev, curr) => prev.concat(interpret(curr)), []);
        }
        f.tokens.push(m);
        return [f];
    }
    return null;
}
function fromPolynomial(p) {
    var t = [];
    for (var i = 0; i < p.term.length; i++) {
        var a = p.term[i];
        if (i > 0 && a.coeff.value > 0)
            t.push(new Symbol("+", false));
        if (a.coeff.approx || a.coeff.value != 1 || Object.keys(a.exponent).length == 0)
            t.push(fromNumeric(a.coeff));
        for (var x in a.exponent) {
            t.push(new Symbol(x, true));
            if (a.exponent[x] != 1) {
                var ex = new Structure(null, StructType.Power);
                ex.elems[0].tokens.push(new Num(a.exponent[x].toString()));
                t.push(ex);
            }
        }
    }
    if (t.length == 0)
        t.push(fromNumeric(Numeric.Zero));
    return t;
}
function fromNumeric(n) {
    if (!n.approx) {
        if (n.den == 1)
            return new Num(n.num.toString());
        var s = new Structure(null, StructType.Frac);
        s.elems[0] = new Formula(s);
        s.elems[0].tokens.push(new Num(n.num.toString()));
        s.elems[1] = new Formula(s);
        s.elems[1].tokens.push(new Num(n.den.toString()));
        return s;
    }
    else {
        return new Num(n.toString());
    }
}
let operator = [
    { symbol: "mod", type: OperatorType.Infix, priority: 0 },
    { symbol: "+", type: OperatorType.Infix, priority: 1 },
    { symbol: "-", type: OperatorType.Infix, priority: 1 },
    { symbol: "+", type: OperatorType.Prefix, priority: 1 },
    { symbol: "-", type: OperatorType.Prefix, priority: 1 },
    { symbol: "*", type: OperatorType.Infix, priority: 2 },
    { symbol: "/", type: OperatorType.Infix, priority: 2 },
    { symbol: "arccos", type: OperatorType.Prefix, priority: 3, func: Math.acos },
    { symbol: "arcsin", type: OperatorType.Prefix, priority: 3, func: Math.asin },
    { symbol: "arctan", type: OperatorType.Prefix, priority: 3, func: Math.atan },
    //{ symbol: "arg", type: OperatorType.Prefix, priority: 3 },
    { symbol: "cos", type: OperatorType.Prefix, priority: 3, func: Math.cos },
    { symbol: "cosh", type: OperatorType.Prefix, priority: 3, func: MathEx.cosh },
    { symbol: "cot", type: OperatorType.Prefix, priority: 3, func: MathEx.cot },
    { symbol: "coth", type: OperatorType.Prefix, priority: 3, func: MathEx.coth },
    { symbol: "csc", type: OperatorType.Prefix, priority: 3, func: MathEx.csc },
    //{ symbol: "det", type: OperatorType.Prefix, priority: 3 },
    { symbol: "exp", type: OperatorType.Prefix, priority: 3, func: Math.exp },
    //{ symbol: "gcd", type: OperatorType.Prefix, priority: 3 },
    { symbol: "lg", type: OperatorType.Prefix, priority: 3, func: MathEx.lg },
    { symbol: "ln", type: OperatorType.Prefix, priority: 3, func: Math.log },
    { symbol: "log", type: OperatorType.Prefix, priority: 3, func: Math.log },
    //{ symbol: "max", type: OperatorType.Prefix, priority: 3 },
    //{ symbol: "min", type: OperatorType.Prefix, priority: 3 },
    { symbol: "sec", type: OperatorType.Prefix, priority: 3, func: MathEx.sec },
    { symbol: "sin", type: OperatorType.Prefix, priority: 3, func: Math.sin },
    { symbol: "sinh", type: OperatorType.Prefix, priority: 3, func: MathEx.sinh },
    { symbol: "tan", type: OperatorType.Prefix, priority: 3, func: Math.tan },
    { symbol: "tanh", type: OperatorType.Prefix, priority: 3, func: MathEx.tanh },
    { symbol: "^", type: OperatorType.Infix, priority: 4 },
    { symbol: "!", type: OperatorType.Suffix, priority: 5 },
    { symbol: "(", type: OperatorType.Prefix, priority: Number.POSITIVE_INFINITY },
    { symbol: "[", type: OperatorType.Prefix, priority: Number.POSITIVE_INFINITY },
    { symbol: "{", type: OperatorType.Prefix, priority: Number.POSITIVE_INFINITY },
    { symbol: ")", type: OperatorType.Suffix, priority: Number.POSITIVE_INFINITY },
    { symbol: "]", type: OperatorType.Suffix, priority: Number.POSITIVE_INFINITY },
    { symbol: "}", type: OperatorType.Suffix, priority: Number.POSITIVE_INFINITY }
];
function getOperator(symbol, type) {
    for (var i = 0; i < operator.length; i++) {
        var o = operator[i];
        if (o.symbol == symbol && o.type == type)
            return o;
    }
    return null;
}
function getPriority(symbol, type) {
    var o = getOperator(symbol, type);
    return o !== null ? o.priority : -1;
}
function evalSeq(t) {
    var opr = operator.map(o => o.symbol);
    var q = [];
    for (var i = 0; i < t.length; i++) {
        var r = null;
        if (t[i] instanceof Symbol) {
            var v = t[i];
            if (opr.indexOf(v.str) < 0)
                r = Polynomial.fromSymbol(v);
        }
        else if (t[i] instanceof Num) {
            var n = t[i];
            if (n.value.indexOf(".") >= 0) {
                r = Numeric.fromReal(parseFloat(n.value));
            }
            else
                r = new Numeric(parseInt(n.value), 1);
        }
        else if (t[i] instanceof Matrix) {
            var m = t[i];
            var x = new Matrix(null, m.rows, m.cols);
            for (var j = 0; j < m.elems.length; j++) {
                var f = new Formula(x);
                f.tokens[0] = evalSeq(m.elems[j].tokens);
                if (f.tokens[0] == null) {
                    x = null;
                    break;
                }
                x.elems[j] = f;
            }
            r = x;
        }
        else if (t[i] instanceof Structure) {
            var s = t[i];
            switch (s.type) {
                case StructType.Frac:
                    var num = evalSeq(s.elems[0].tokens);
                    var den = evalSeq(s.elems[1].tokens);
                    r = mul(num, den, true);
                    break;
            }
        }
        else if (t[i] instanceof Formula) {
            var f = t[i];
            r = evalSeq(f.tokens);
            if (f.prefix == "√" && f.suffix == "")
                r = realFunc(r, Math.sqrt);
            else if (f.prefix == "|" && f.suffix == "|"
                || f.prefix == "‖" && f.suffix == "‖")
                r = norm(r);
            else if (f.prefix == "⌊" && f.suffix == "⌋")
                r = floor(r);
            else if (f.prefix == "⌈" && f.suffix == "⌉")
                r = ceil(r);
        }
        q[i] = (r !== null ? r : t[i]);
    }
    return evalSeqMain(q, 0, 0);
}
function evalSeqMain(t, index, border) {
    var res = null;
    var argc = 0;
    console.debug("evalSeqMain " + t.toString() + " " + index + " " + border);
    if (t.length == 0 || t.length <= index)
        return null;
    else if (t.length == 1)
        return t[0];
    if (t[index] instanceof Symbol) {
        console.debug("eval 0 symbol");
        var opr = getOperator(t[index].str, OperatorType.Prefix);
        if (opr != null && opr.priority >= border) {
            if (opr.symbol == "(" || opr.symbol == "[" || opr.symbol == "{") {
                argc = 3;
                evalSeqMain(t, index + 1, 0);
                console.debug("eval br " + t.toString());
                res = t[index + 1];
            }
            else {
                argc = 2;
                evalSeqMain(t, index + 1, opr.priority + 1);
                if (opr.symbol == "-")
                    res = negate(t[index + 1]);
                else if (opr.symbol == "+")
                    res = t[index + 1];
                else if (opr.func)
                    res = realFunc(t[index + 1], opr.func);
            }
        }
    }
    else if (t[index + 1] instanceof Symbol) {
        console.debug("eval 1 symbol");
        var o = t[index + 1];
        var p;
        if ((p = getPriority(o.str, OperatorType.Infix)) >= border) {
            argc = 3;
            evalSeqMain(t, index + 2, p + 1);
            if (o.str == "+")
                res = add(t[index], t[index + 2], false);
            else if (o.str == "-")
                res = add(t[index], t[index + 2], true);
            else if (o.str == "*" || o.str == "·" || o.str == "∙")
                res = mul(t[index], t[index + 2], false);
            else if (o.str == "/" || o.str == "÷")
                res = mul(t[index], t[index + 2], true);
            else if (o.str == "mod")
                res = mod(t[index], t[index + 2]);
        }
        else if ((p = getPriority(o.str, OperatorType.Suffix)) >= border) {
            argc = 2;
            evalSeqMain(t, index + 2, p + 1);
            if (o.str == "!")
                res = factorial(t[index]);
        }
        else if ((p = getPriority(o.str, OperatorType.Prefix)) >= border) {
            argc = 2;
            evalSeqMain(t, index + 1, 0);
            res = mul(t[index], t[index + 1], false);
        }
    }
    else if (t[index + 1] instanceof Structure
        && t[index + 1].type == StructType.Power) {
        console.debug("eval 1 pow");
        var res = null;
        var s = t[index + 1];
        var p = getPriority("^", OperatorType.Infix);
        if (p >= border) {
            argc = 2;
            res = power(t[index], evalSeq(s.elems[0].tokens.slice(0)));
        }
    }
    else if (t.length - index >= 2) {
        console.debug("eval 0 mul");
        var p = getPriority("*", OperatorType.Infix);
        if (p >= border) {
            argc = 2;
            evalSeqMain(t, index + 1, p + 1);
            res = mul(t[index], t[index + 1], false);
        }
    }
    if (res !== null) {
        t.splice(index, argc, res);
        console.debug("eval res " + t.toString() + " " + index + " " + border);
        if (t.length - index > 1)
            evalSeqMain(t, index, 0);
    }
    else {
        console.debug("eval none");
    }
    return t.length == 1 ? t[0] : null;
}
function add(x, y, sub) {
    if (x instanceof Formula && x.tokens.length == 1)
        x = x.tokens[0];
    if (y instanceof Formula && y.tokens.length == 1)
        y = y.tokens[0];
    console.debug((sub ? "sub" : "add") + " " + (x ? x.toString() : "?") + " " + (y ? y.toString() : "?"));
    if (x instanceof Numeric && y instanceof Numeric) {
        var m = x;
        var n = y;
        return sub ? Numeric.sub(m, n) : Numeric.add(m, n);
    }
    else if (x instanceof Polynomial || y instanceof Polynomial) {
        var p, q;
        if (x instanceof Polynomial)
            p = x;
        else if (x instanceof Numeric)
            p = Polynomial.fromNumeric(x);
        if (y instanceof Polynomial)
            q = y;
        else if (y instanceof Numeric)
            q = Polynomial.fromNumeric(y);
        return sub ? Polynomial.sub(p, q) : Polynomial.add(p, q);
    }
    else if (x instanceof Matrix && y instanceof Matrix) {
        var a = x;
        var b = y;
        if (a.rows == b.rows && a.cols == b.cols) {
            var r = new Matrix(null, a.rows, a.cols);
            for (var i = 0; i < a.count(); i++) {
                var s = add(a.elems[i], b.elems[i], sub);
                if (s === null)
                    return null;
                r.elems[i] = new Formula(r);
                r.elems[i].tokens.push(s);
            }
        }
        return r;
    }
    return null;
}
function mul(x, y, div) {
    if (x instanceof Formula && x.tokens.length == 1)
        x = x.tokens[0];
    if (y instanceof Formula && y.tokens.length == 1)
        y = y.tokens[0];
    console.debug((div ? "div" : "mul") + " " + (x ? x.toString() : "?") + " " + (y ? y.toString() : "?"));
    if (x instanceof Numeric && y instanceof Numeric) {
        var m = x;
        var n = y;
        return div ? Numeric.div(m, n) : Numeric.mul(m, n);
    }
    else if (x instanceof Numeric && y instanceof Matrix && !div) {
        var b = y;
        var r = new Matrix(null, b.rows, b.cols);
        for (var i = 0; i < b.rows; i++) {
            for (var j = 0; j < b.cols; j++) {
                r.elems[i * b.cols + j] = new Formula(r);
                var s = mul(x, b.tokenAt(i, j), false);
                if (s === null)
                    return null;
                r.elems[i * b.cols + j].tokens.push(s);
            }
        }
        return r;
    }
    else if (x instanceof Polynomial || y instanceof Polynomial) {
        var p, q;
        if (x instanceof Polynomial)
            p = x;
        else if (x instanceof Numeric)
            p = Polynomial.fromNumeric(x);
        if (y instanceof Polynomial)
            q = y;
        else if (y instanceof Numeric)
            q = Polynomial.fromNumeric(y);
        return div ? Polynomial.div(p, q) : Polynomial.mul(p, q);
    }
    else if (x instanceof Matrix && y instanceof Matrix) {
        var a = x;
        var b = y;
        if (a.cols == b.rows) {
            var r = new Matrix(null, a.rows, b.cols);
            for (var i = 0; i < a.rows; i++) {
                for (var j = 0; j < b.cols; j++) {
                    r.elems[i * a.cols + j] = new Formula(r);
                    var s = mul(a.tokenAt(i, 0), b.tokenAt(0, j), false);
                    for (var k = 1; k < a.cols; k++) {
                        s = add(s, mul(a.tokenAt(i, k), b.tokenAt(k, j), false), false);
                        if (s == null)
                            return null;
                    }
                    r.elems[i * b.cols + j].tokens.push(s);
                }
            }
        }
        return r;
    }
    return null;
}
function mod(x, y) {
    if (x instanceof Formula && x.tokens.length == 1)
        x = x.tokens[0];
    if (y instanceof Formula && y.tokens.length == 1)
        y = y.tokens[0];
    if (x instanceof Numeric && y instanceof Numeric) {
        var m = x;
        var n = y;
        if (m.den == 1 && n.den == 1) {
            var r = m.num % n.num;
            if (r < 0)
                r += n.num;
            return new Numeric(r, 1, m.approx || n.approx);
        }
    }
    return null;
}
function negate(x) {
    if (x instanceof Numeric) {
        return Numeric.negate(x);
    }
    else if (x instanceof Polynomial) {
        return Polynomial.negate(x);
    }
    else if (x instanceof Matrix) {
        var a = x;
        var r = new Matrix(null, a.rows, a.cols);
        for (var i = 0; i < a.rows; i++) {
            for (var j = 0; j < a.cols; j++) {
                r.elems[i * a.cols + j] = new Formula(r);
                var s = add(Numeric.Zero, a.tokenAt(i, j), true);
                if (s === null)
                    return null;
                r.elems[i * a.cols + j].tokens.push(s);
            }
        }
        return r;
    }
    return null;
}
function power(x, y) {
    if (x instanceof Numeric && y instanceof Numeric) {
        var a = x;
        var b = y;
        if (!a.approx && !b.approx && b.den == 1) {
            var ex = Math.abs(b.num);
            var p = Math.pow(a.num, ex);
            var q = Math.pow(a.den, ex);
            return b.num >= 0 ? new Numeric(p, q) : new Numeric(q, p);
        }
        return Numeric.fromReal(Math.pow(a.value, b.value));
    }
    else if (x instanceof Polynomial && y instanceof Numeric) {
        var f = x;
        var b = y;
        if (!b.approx && b.num > 0 && b.den == 1) {
            var ex = Math.abs(b.num);
            var g = f;
            for (var i = 1; i < ex; i++)
                g = Polynomial.mul(f, g);
            return g;
        }
    }
    else if (x instanceof Matrix && y instanceof Numeric) {
        var m = x;
        var b = y;
        if (!b.approx && b.num > 0 && b.den == 1) {
            var ex = Math.abs(b.num);
            var r = m;
            for (var i = 1; i < ex; i++)
                r = mul(r, m, false);
            return r;
        }
    }
    return null;
}
function realFunc(x, f) {
    if (x instanceof Numeric) {
        var m = x;
        return Numeric.fromReal(f(m.value));
    }
    return null;
}
function factorial(x) {
    if (x instanceof Numeric) {
        var m = x;
        if (!m.approx && m.num >= 0 && m.den == 1) {
            var r = 1;
            for (var i = 2; i <= m.num; i++)
                r *= i;
            return new Numeric(r, 1, false);
        }
    }
    return null;
}
function norm(x) {
    if (x instanceof Numeric) {
        var m = x;
        return new Numeric(Math.abs(m.num), m.den, m.approx);
    }
    return null;
}
function floor(x) {
    if (x instanceof Numeric) {
        var m = x;
        return new Numeric(Math.floor(m.value), 1, m.approx);
    }
    return null;
}
function ceil(x) {
    if (x instanceof Numeric) {
        var m = x;
        return new Numeric(Math.ceil(m.value), 1, m.approx);
    }
    return null;
}

class Rect {
    constructor(left, top, width, height) {
        this.left = left;
        this.top = top;
        this.right = left + width;
        this.bottom = top + height;
        this.width = width;
        this.height = height;
    }
    static fromJQuery(a) {
        var pos = a.offset();
        return new Rect(pos.left, pos.top, a.width(), a.height());
    }
    center() {
        return { x: (this.left + this.right) / 2, y: (this.top + this.bottom) / 2 };
    }
    size() {
        return { width: this.width, height: this.height };
    }
    contains(a) {
        if (a instanceof Rect) {
            var r = a;
            return r.left >= this.left && r.right <= this.right && r.top >= this.top && r.bottom <= this.bottom;
        }
        else {
            var p = a;
            return p.x >= this.left && p.x <= this.right && p.y >= this.top && p.y <= this.bottom;
        }
    }
}
var InputType;
(function (InputType) {
    InputType[InputType["Empty"] = 0] = "Empty";
    InputType[InputType["Number"] = 1] = "Number";
    InputType[InputType["Symbol"] = 2] = "Symbol";
    InputType[InputType["String"] = 3] = "String";
})(InputType || (InputType = {}));
var RecordType;
(function (RecordType) {
    RecordType[RecordType["Transfer"] = 0] = "Transfer";
    RecordType[RecordType["Edit"] = 1] = "Edit";
    RecordType[RecordType["EditMatrix"] = 2] = "EditMatrix";
    RecordType[RecordType["DiagramEdit"] = 3] = "DiagramEdit";
    RecordType[RecordType["DiagramDeco"] = 4] = "DiagramDeco";
})(RecordType || (RecordType = {}));
class TypeMath {
    /////////////////
    /* public func */
    /////////////////
    constructor($field, $latex, $candy, $ghost, $selectedArea, $debug) {
        this.$field = $field;
        this.$latex = $latex;
        this.$candy = $candy;
        this.$ghost = $ghost;
        this.$selectedArea = $selectedArea;
        this.$debug = $debug;
        this.candMax = 16;
        this._logText = "";
        this.formula = new Formula(null);
        this.activeField = this.formula;
        this.activeIndex = 0;
        this.markedIndex = -1;
        this.candIndex = -1;
        this.candCount = 0;
        this.candSelected = "";
        this.currentInput = "";
        this.postInput = "";
        this.inputType = InputType.Empty;
        this.inputEscaped = false;
        this.diagramOption = {
            from: -1,
            to: -1,
            arrowIndex: -1,
            num: 1,
            style: StrokeStyle.Plain,
            head: ">"
        };
        this.macroOption = {
            field: null,
            epoch: 0
        };
        this.clipboard = [];
        this.records = [];
        this.dragFrom = null;
        this.dragRect = null;
        this.digits = ["0", "1", "2", "3", "4", "5", "6", "7", "8", "9"];
        this.symbols = [
            "+", "-", "*", "/", "^", "_", "!", ":", ",", "<=", ">=", "(", ")", "[", "]", "{", "}", "|"
        ];
        this.operators = [
            "∑", "∏", "∐", "⋂", "⋃", "⨄", "⨆", "⋁", "⋀", "⨁", "⨂", "⨀",
            "∫", "∮", "∬", "∭", "⨌"
        ];
        this.keywords = {
            "and": "∧",
            "or": "∨",
            "not": "¬",
            "imp": "→",
            "<=": "≤",
            ">=": "≥",
            "infer": "",
            "frac": "",
            "matrix": "",
            "pmatrix": "(",
            "bmatrix": "[",
            "Bmatrix": "{",
            "vmatrix": "|",
            "Vmatrix": "‖",
            "(": "(",
            "[": "[",
            "|": "|",
            "floor": "⌊",
            "ceil": "⌈",
            "angle": "〈",
            "sqrt": "√",
            "xymatrix": "",
        };
        this.macroTable = {};
        this.bracketCor = {
            "(": ")", "{": "}", "[": "]", "|": "|", "‖": "‖", "⌊": "⌋", "⌈": "⌉", "〈": "〉", "√": ""
        };
        this.enrichKeywords();
        this.$latex.focus();
        this.$latex.on('keydown', (e) => { this.processInput(e.originalEvent); });
        this.$ghost.on('mousedown', (e) => { this.dragFrom = { x: e.pageX, y: e.pageY }; this.jumpTo(this.dragFrom); });
        this.$ghost.on('mousemove', (e) => { this.dragSelect(e); });
        this.$ghost.on('mouseup', (e) => { this.dragFrom = this.dragRect = null; this.render(); this.$latex.focus(); });
        this._glyph = new GlyphFactory();
        this.render();
    }
    get selectingArrow() {
        return this.activeField instanceof Diagram
            && this.diagramOption.arrowIndex >= 0;
    }
    get inMacroMode() {
        return this.macroOption.field !== null;
    }
    getLaTeXCode() {
        let macrocode = "";
        for (var macro in this.macroTable) {
            var m = this.macroTable[macro];
            var s = "\\newcommand{\\" + macro + "}";
            if (m.argc > 0)
                s += "[" + m.argc.toString() + "]";
            s += "{" + trans(m.content) + "}";
            macrocode += s + "\n";
        }
        if (macrocode.length > 0)
            macrocode += "\n";
        return macrocode + trans(this.formula, "", this.proofMode);
    }
    setLaTeXCode(src) {
        // clear formula
        this.$active = null;
        this.formula = new Formula(null);
        this.activeField = this.formula;
        this.activeIndex = 0;
        var code = LaTeXReader.parse(src);
        this.interpretLaTeX(code);
        this.render();
    }
    //////////////////
    /* private func */
    //////////////////
    enrichKeywords() {
        var dic = [symbols, accentSymbols];
        for (var i = 0; i < dic.length; i++) {
            for (var c in dic[i]) {
                var key = dic[i][c];
                if (!(key in this.keywords))
                    this.keywords[key] = c;
            }
        }
        for (var c in styles)
            this.keywords[c] = "";
    }
    render() {
        var a = (this.currentInput != "" ? this.currentInput : SixPerEmSpace) + this.postInput;
        if (this.inputEscaped)
            a = "\\" + a;
        this.$active = null;
        this.afterLayout = [];
        this.$field.empty();
        this.outputCurrentStyle = [FontStyle.Normal];
        this.outputToken(this.$field, this.formula, { arguments: null, actGenerated: false });
        if (this.$active) {
            this.$active.text(a);
            a = null;
        }
        this.drawAfterLayout();
        if (a && this.$active)
            this.$active.text(a);
        this.showCandidate();
        if (this.dragFrom && this.dragRect) {
            var ofs = this.$field.offset();
            this.$selectedArea.css({
                visibility: "visible",
                left: (this.dragRect.left - ofs.left) + "px",
                top: (this.dragRect.top - ofs.top) + "px",
                width: this.dragRect.width + "px",
                height: this.dragRect.height + "px"
            });
        }
        else
            this.$selectedArea.css("visibility", "hidden");
        this.$latex.val(this.getLaTeXCode());
        if (this.$debug)
            this.$debug.text([
                "formula       = " + this.formula.toString(),
                "activeFormula = " + this.activeField.toString(),
                "activeIndex   = " + this.activeIndex.toString(),
                "markedIndex   = " + this.markedIndex.toString(),
                "candIndex     = " + this.candIndex.toString(),
                "candCount     = " + this.candCount.toString(),
                "candSelected  = " + this.candSelected.toString(),
                "currentInput  = " + this.currentInput.toString(),
                "inputType     = " + (this.inputType == InputType.Empty ? "Empty" :
                    this.inputType == InputType.Number ? "Number" :
                        this.inputType == InputType.String ? "String" : "Symbol"),
                "diag.from     = " + this.diagramOption.from,
                "diag.to       = " + this.diagramOption.to,
                "diag.arrow    = " + this.diagramOption.arrowIndex,
                "diag.num      = " + this.diagramOption.num,
                "diag.style    = " + this.diagramOption.style,
                "diag.head     = " + this.diagramOption.head,
                "macro.field   = " + (this.macroOption.field ? this.macroOption.field.toString() : "null"),
                "macro.epoch   = " + this.macroOption.epoch,
                "macroTable    = " + Object.keys(this.macroTable).join(", "),
                "clipboard     = " + this.clipboard.toString(),
                "records       = " + this.records.map(this.writeRecord).join("\n")
            ].join("\n"));
    }
    writeRecord(r) {
        return Object.keys(r).map(p => p + ":" + r[p]).join(", ");
    }
    //////////////////////////////////////
    /*  event handling					*/
    //////////////////////////////////////
    processInput(e) {
        var key = knowKey(e);
        if (key == "") {
            this.processControlInput(e);
            return;
        }
        else if (e.ctrlKey) {
            this.processModifiedInput(e, key);
            return;
        }
        this.markedIndex = -1;
        if (this.activeField instanceof Diagram
            && this.processInputDiagram(key)) {
            this.render();
            if (this.changed)
                this.changed();
            e.preventDefault();
            return;
        }
        if (!(this.activeField instanceof Formula))
            this.enterFormula(true);
        var t = this.getInputType(key);
        if (this.inputType == InputType.Empty) {
            if (key == '\\')
                this.inputEscaped = true;
            else
                this.currentInput += key;
            this.inputType = t;
        }
        else {
            if (t == this.inputType ||
                (this.inMacroMode
                    && this.currentInput.charAt(0) == '#'
                    && t == InputType.Number)) {
                this.currentInput += key;
            }
            else {
                this.interpretInput();
                this.inputType = t;
                this.currentInput += key;
            }
        }
        this.render();
        if (this.changed)
            this.changed();
        e.preventDefault();
    }
    getInputType(s) {
        if (this.digits.indexOf(s) >= 0
            || this.inputType == InputType.Number && s == ".")
            return InputType.Number;
        else if (this.symbols.indexOf(s) >= 0)
            return InputType.Symbol;
        else
            return InputType.String;
    }
    processInputDiagram(key) {
        var processed = true;
        if (this.diagramOption.from < 0) {
            processed = this.decorateObject(key);
        }
        else if (this.selectingArrow) {
            switch (key) {
                case "^":
                    this.labelArrow(LabelPosotion.Left);
                    break;
                case "|":
                    this.labelArrow(LabelPosotion.Middle);
                    break;
                case "_":
                    this.labelArrow(LabelPosotion.Right);
                    break;
            }
        }
        else {
            switch (key) {
                case "n":
                    if (this.diagramOption.style == StrokeStyle.None) {
                        this.diagramOption.style = StrokeStyle.Plain;
                        this.diagramOption.head = ">";
                    }
                    else {
                        this.diagramOption.style = StrokeStyle.None;
                        this.diagramOption.head = "";
                    }
                    break;
                case "=":
                    this.diagramOption.num = 2;
                case "-":
                case "d":
                    this.diagramOption.style =
                        (this.diagramOption.style == StrokeStyle.Plain
                            ? StrokeStyle.Dashed : StrokeStyle.Plain);
                    break;
                case ":":
                    this.diagramOption.num = 2;
                case ".":
                    this.diagramOption.style = StrokeStyle.Dotted;
                    break;
                case ":":
                    this.diagramOption.num = 2;
                case ".":
                    this.diagramOption.style = StrokeStyle.Dotted;
                    break;
                case "~":
                    this.diagramOption.style = StrokeStyle.Wavy;
                    break;
                case "1":
                    this.diagramOption.num = 1;
                    break;
                case "2":
                    this.diagramOption.num = 2;
                    break;
                case "3":
                    this.diagramOption.num = 3;
                    break;
                case "4":
                    this.diagramOption.num = 4;
                    break;
                default:
                    processed = false;
            }
        }
        return processed;
    }
    processControlInput(e) {
        var suppress = true;
        var key = knowControlKey(e);
        switch (key) {
            case ControlKey.Tab:
                if (this.candIndex >= 0)
                    this.currentInput = this.candSelected;
                else if (this.selectingArrow)
                    this.diagramOption.arrowIndex++; // moduloed in drawArrows()
                break;
            case ControlKey.Enter:
                if (this.candIndex >= 0) {
                    this.decideCandidate();
                    break;
                }
                if (this.activeField instanceof Formula) {
                    var f = this.activeField;
                    if (this.activeIndex > 0
                        && f.tokens[this.activeIndex - 1] instanceof Symbol
                        && f.tokens[this.activeIndex - 1].str == "=") {
                        var res = evalToken(f.tokens.slice(0, this.activeIndex - 1));
                        this.pasteToken(res !== null ? res : [new Symbol("?", false)]);
                        break;
                    }
                }
                if (this.activeField.parent instanceof Diagram)
                    this.leaveFormula(true);
                else if (this.activeField instanceof Diagram) {
                    if (this.diagramOption.from >= 0)
                        this.addArrow();
                    else
                        this.enterFormula(true);
                }
                break;
            case ControlKey.Space:
                if (this.currentInput == "") {
                    if (e.shiftKey) {
                        this.markedIndex = -1;
                        this.movePrev();
                    }
                    else {
                        if (this.activeField.parent instanceof Structure) {
                            var p = this.activeField.parent;
                            if (p.type == StructType.Infer
                                && p.elems[1] == this.activeField) {
                                this.inputType = InputType.String;
                                this.currentInput = "&";
                                this.interpretInput();
                                break;
                            }
                        }
                        else if (this.activeField instanceof Diagram) {
                            this.diagramOption.from = (this.diagramOption.from < 0 ? this.activeIndex : -1);
                            break;
                        }
                        this.moveNext();
                    }
                }
                else
                    this.interpretInput();
                break;
            case ControlKey.Left:
            case ControlKey.Right:
                if (e.ctrlKey)
                    this.modifyMatrix(true, key == ControlKey.Right);
                else
                    this.moveHorizontal(key == ControlKey.Right);
                break;
            case ControlKey.Up:
            case ControlKey.Down:
                if (e.ctrlKey)
                    this.modifyMatrix(false, key == ControlKey.Down);
                else if (this.currentInput != "")
                    this.changeCandidate(key == ControlKey.Down);
                else
                    this.moveVertical(key == ControlKey.Up);
                break;
            case ControlKey.Backspace:
                if (this.currentInput != "") {
                    this.currentInput = this.currentInput.slice(0, -1);
                    if (this.currentInput == "")
                        this.inputType = InputType.Empty;
                }
                else if (this.markedIndex >= 0) {
                    this.removeToken(this.markedIndex, this.activeIndex);
                    this.markedIndex = -1;
                }
                else if (this.selectingArrow)
                    this.removeArrow();
                else if (this.activeIndex > 0) {
                    this.removeToken(this.activeIndex - (this.activeField instanceof Structure ? 0 : 1), this.activeIndex);
                }
                break;
            case ControlKey.Shift:
                if (this.markedIndex < 0)
                    this.markedIndex = this.activeIndex;
                else
                    this.markedIndex = -1;
                break;
            default:
                suppress = false;
        }
        if (suppress)
            e.preventDefault();
        this.render();
        if (this.changed)
            this.changed();
    }
    processModifiedInput(e, key) {
        var suppress = true;
        switch (key) {
            case "c": // clipboard copy
            case "x":
                if (this.markedIndex >= 0) {
                    this.clipboard = this.activeField.copy(this.markedIndex, this.activeIndex);
                    if (key == "x")
                        this.removeToken(this.markedIndex, this.activeIndex);
                    this.markedIndex = -1;
                }
                break;
            case "v":
                if (this.clipboard != null)
                    this.pasteToken(this.clipboard);
                break;
            case "z":
                this.undo();
                break;
            case "m":
                if (!this.inMacroMode)
                    this.enterMacroMode();
                else
                    this.exitMacroMode();
                break;
            default:
                suppress = false;
                break;
        }
        if (suppress)
            e.preventDefault();
        this.render();
        if (this.changed)
            this.changed();
    }
    //////////////////////////////////////
    /*  Import LaTeX                    */
    //////////////////////////////////////
    interpretLaTeX(code) {
        switch (code.type) {
            case LaTeXASTType.Sequence:
                console.debug("LaTeX: seq");
                for (var i = 0; i < code.children.length; i++)
                    this.interpretLaTeX(code.children[i]);
                break;
            case LaTeXASTType.Environment:
            case LaTeXASTType.Command:
                console.debug("LaTeX: env/cmd " + code.value);
                if (code.value == "newcommand") {
                    this.enterMacroMode();
                    this.interpretLaTeX(code.children[2]);
                    this.registerMacro(code.children[0].value, this.macroOption.field.tokens.map(t => t.clone(null)));
                    this.exitMacroMode(false, true);
                }
                else if (code.value == "left") {
                    this.interpretLaTeXCode(code.children[0].value, InputType.Symbol, true);
                }
                else if (code.value == "right") {
                    this.moveNext();
                }
                else if (code.value == "infer") {
                    this.interpretLaTeXCode(code.value, InputType.String);
                    for (var i = 1; i <= 2; i++) {
                        this.interpretLaTeX(code.children[i]);
                        this.moveNext();
                    }
                    if (code.children[0])
                        this.interpretLaTeX(code.children[0]);
                    this.moveNext();
                }
                else if (code.value == "*") {
                    if (!(this.activeField instanceof Diagram)
                        && this.activeField.parent instanceof Diagram)
                        this.leaveFormula(true);
                    else
                        break;
                    if (code.children[2])
                        this.decorateObject("f");
                    var deco;
                    if (code.children[0]) {
                        deco = code.children[0].value;
                        for (var i = 0; i < deco.length; i++)
                            this.decorateObject(deco.charAt(i));
                    }
                    if (code.children[1])
                        this.decorateObject(code.children[1].value);
                    if (code.children[2]) {
                        switch (code.children[2].value) {
                            case "F--":
                                this.decorateObject("d");
                                break;
                            case "F.":
                                this.decorateObject(".");
                                break;
                            case "F=":
                                this.decorateObject("=");
                                break;
                        }
                    }
                    this.enterFormula(true);
                    this.interpretLaTeX(code.children[3]);
                }
                else if (code.value == "ar") {
                    if (!(this.activeField instanceof Diagram)
                        && this.activeField.parent instanceof Diagram)
                        this.leaveFormula(true);
                    else
                        break;
                    var num = 1;
                    var style = StrokeStyle.Plain;
                    var dir = 0;
                    if (code.children[0])
                        num = parseInt(code.children[0].value);
                    if (code.children[1]) {
                        switch (code.children[1].value) {
                            case "=>": num = 2;
                            case "->":
                                style = StrokeStyle.Plain;
                                break;
                            case "==>": num = 2;
                            case "-->":
                                style = StrokeStyle.Dashed;
                                break;
                            case ":>": num = 2;
                            case ".>":
                                style = StrokeStyle.Dotted;
                                break;
                            case "~>":
                                style = StrokeStyle.Wavy;
                                break;
                            case "":
                                style = StrokeStyle.None;
                                this.diagramOption.head = "";
                                break;
                        }
                    }
                    if (code.children[2]) {
                        var cols = this.activeField.cols;
                        var d = code.children[2].value;
                        for (var i = 0; i < d.length; i++) {
                            switch (d.charAt(i)) {
                                case "l":
                                    dir--;
                                    break;
                                case "r":
                                    dir++;
                                    break;
                                case "u":
                                    dir -= cols;
                                    break;
                                case "d":
                                    dir += cols;
                                    break;
                            }
                        }
                    }
                    this.diagramOption.num = num;
                    this.diagramOption.from = this.activeIndex;
                    this.diagramOption.style = style;
                    this.activeIndex += dir;
                    this.addArrow();
                    if (code.children[3]) {
                        var pos = LabelPosotion.Left;
                        if (code.children[3].value == "|")
                            pos = LabelPosotion.Middle;
                        else if (code.children[3].value == "_")
                            pos = LabelPosotion.Right;
                        this.diagramOption.from = this.activeIndex;
                        this.activeIndex += dir;
                        var n = this.activeField
                            .countArrow(this.diagramOption.from, this.activeIndex);
                        this.diagramOption.arrowIndex = n - 1;
                        this.labelArrow(pos);
                        this.interpretLaTeX(code.children[4]);
                        this.leaveFormula(false);
                    }
                    this.enterFormula(false);
                }
                else if (code.value.indexOf("matrix") >= 1) {
                    this.interpretLaTeXCode(code.value, InputType.String);
                    var cols = code.children[0].children.length;
                    for (var i = 1; i < code.children.length; i++) {
                        this.modifyMatrix(false, true);
                        cols = Math.max(cols, code.children[i].children.length);
                    }
                    for (var j = 1; j < cols; j++)
                        this.modifyMatrix(true, true);
                    if (code.value == "xymatrix")
                        this.enterFormula(true);
                    for (var i = 0; i < code.children.length; i++) {
                        var row = code.children[i];
                        for (var j = 0; j < row.children.length; j++) {
                            var cell = row.children[j];
                            this.interpretLaTeX(cell);
                            this.moveNext();
                        }
                    }
                    this.moveNext();
                }
                else {
                    this.interpretLaTeXCode(code.value, this.symbols.indexOf(code.value) >= 0
                        ? InputType.Symbol : InputType.String, true);
                    for (var i = 0; i < code.children.length; i++) {
                        this.interpretLaTeX(code.children[i]);
                        this.moveNext();
                    }
                }
                break;
            case LaTeXASTType.Number:
                console.debug("LaTeX: n " + code.value);
                this.interpretLaTeXCode(code.value, InputType.Number);
                break;
            case LaTeXASTType.Symbol:
                console.debug("LaTeX: s " + code.value);
                this.interpretLaTeXCode(code.value, this.symbols.indexOf(code.value) >= 0
                    ? InputType.Symbol : InputType.String);
                break;
        }
    }
    interpretLaTeXCode(code, type, force = false) {
        this.inputType = type;
        this.currentInput = code;
        this.interpretInput(force, false);
    }
    //////////////////////////////////////
    /*  macro registering				*/
    //////////////////////////////////////
    enterMacroMode() {
        var epoch = this.records.length;
        if (this.markedIndex < 0)
            this.insertToken(new Formula(this.activeField));
        else {
            this.extrudeToken(this.markedIndex, this.activeIndex);
            this.markedIndex = -1;
        }
        this.enterFormula(true);
        this.macroOption.field = this.activeField;
        this.macroOption.epoch = epoch;
    }
    exitMacroMode(register = true, force = false) {
        if (register) {
            var name;
            for (var accepted = false; !accepted;) {
                name = window.prompt("Macro Name:");
                if (name === null)
                    return false;
                else if (name in this.keywords)
                    window.alert("Duplicate Name!");
                else if (name.length == 0)
                    window.alert("No Input!");
                else
                    accepted = true;
            }
            this.registerMacro(name, this.macroOption.field.tokens.map(t => t.clone(null)));
        }
        else if (!force) {
            if (!window.confirm("Would you exit macro mode?"))
                return false;
        }
        this.macroOption.field = null;
        while (this.records.length > this.macroOption.epoch)
            this.undo();
        this.macroOption.epoch = 0;
        return true;
    }
    registerMacro(name, content) {
        if (name in this.keywords)
            console.error("[Application.registerMacro] duplicate macro name");
        var n = this.countArgs({
            count: () => { return content.length; },
            token: (i) => { return content[i]; }
        });
        var t;
        if (content.length == 1)
            t = content[0];
        else {
            var f = new Formula(null);
            f.tokens = content;
            t = f;
        }
        this.macroTable[name] = { argc: n, content: t };
        this.keywords[name] = "";
    }
    countArgs(seq) {
        var n = seq.count();
        var count = 0;
        for (var i = 0; i < n; i++) {
            var t = seq.token(i);
            var c = 0;
            if (t instanceof Structure || t instanceof Formula) {
                var s = t;
                c = this.countArgs(s);
            }
            else if (t instanceof Symbol) {
                var m = t.str.match(/^#(\d+)$/);
                if (m)
                    c = parseInt(m[1]);
            }
            count = Math.max(count, c);
        }
        return count;
    }
    //////////////////////////////////////
    /*  transition						*/
    //////////////////////////////////////
    movePrev() {
        if (this.activeIndex == 0) {
            if (this.transferFormula(false))
                return;
        }
        this.moveHorizontal(false);
    }
    moveNext() {
        if (this.activeIndex == this.activeField.count()) {
            if (this.transferFormula(true))
                return;
        }
        this.moveHorizontal(true);
    }
    moveHorizontal(forward) {
        if (this.currentInput != "") {
            if (!forward) {
                this.currentInput = this.currentInput.slice(0, -1);
                return;
            }
            else
                this.interpretInput();
        }
        var dif = forward ? 1 : -1;
        if (this.activeField instanceof Formula) {
            if (this.activeIndex + dif >= 0
                && this.activeIndex + dif <= this.activeField.count()) {
                if (!this.enterFormula(forward))
                    this.activeIndex += dif;
                return;
            }
        }
        else if (this.activeField instanceof Matrix) {
            var m = this.activeField;
            var c = this.activeIndex % m.cols;
            if (forward && c < m.cols - 1 || !forward && c > 0) {
                this.activeIndex += (forward ? 1 : -1);
                return;
            }
        }
        var p = this.activeField.parent;
        if (p == null) {
            return;
        }
        else if (p instanceof Matrix) {
            var m = p;
            var a = m.around.bind(m)(this.activeField, true, forward);
            if (a != null) {
                if (this.markedIndex >= 0) {
                    this.leaveFormula(forward);
                    var c = this.activeIndex % m.cols;
                    if (forward && c < m.cols - 1 || !forward && c > 0)
                        this.activeIndex += (forward ? 1 : -1);
                    return;
                }
                if (this.transferFormula(forward, a))
                    return;
            }
        }
        else if (p instanceof Structure) {
            if (this.transferFormula(forward))
                return;
        }
        this.leaveFormula(forward);
    }
    moveVertical(upward) {
        if (this.activeField instanceof Matrix) {
            var m = this.activeField;
            var r = Math.floor(this.activeIndex / m.cols);
            if (upward && r > 0 || !upward && r < m.rows - 1) {
                this.activeIndex += (upward ? -m.cols : m.cols);
                return;
            }
        }
        var ac = this.activeField;
        var p = ac.parent;
        if (this.markedIndex >= 0 && p instanceof Matrix) {
            this.leaveFormula(true);
            var m = p;
            var r = Math.floor(this.activeIndex / m.cols);
            if (upward && r > 0 || !upward && r < m.rows - 1)
                this.activeIndex += (upward ? -m.cols : m.cols);
            return;
        }
        while (p != null) {
            if (p instanceof Structure) {
                var s = p;
                var neig;
                if (s.type == StructType.Infer)
                    neig = (upward ? s.next : s.prev)(ac);
                else if (s instanceof Matrix)
                    neig = s.around.bind(s)(ac, false, !upward);
                else
                    neig = (upward ? s.prev : s.next)(ac);
                if (neig != null) {
                    this.transferFormula(false, neig);
                    var rect = this.$active[0].getBoundingClientRect();
                    var x0 = (rect.left + rect.right) / 2;
                    var a = [];
                    for (var i = 0; i < neig.tokens.length; i++) {
                        var rect = neig.tokens[i].renderedElem[0].getBoundingClientRect();
                        if (i == 0)
                            a.push(rect.left);
                        a.push(rect.right);
                    }
                    this.activeIndex = (a.length == 0) ? 0
                        : a.map((x, i) => ({ d: Math.abs(x - x0), index: i }))
                            .reduce((prev, curr) => (curr.d < prev.d) ? curr : prev).index;
                    return;
                }
            }
            ac = p;
            p = p.parent;
        }
    }
    //////////////////////////////////////
    /*  autocomplete					*/
    //////////////////////////////////////
    changeCandidate(next) {
        if (this.candIndex < 0)
            return;
        if (next && ++this.candIndex >= this.candCount)
            this.candIndex = 0;
        if (!next && --this.candIndex < 0)
            this.candIndex = this.candCount - 1;
    }
    decideCandidate() {
        if (this.candIndex < 0) {
            this.pushSymbols();
            return;
        }
        this.currentInput = this.candSelected;
        this.interpretInput(true);
    }
    showCandidate() {
        var key = this.currentInput;
        var keys = Object.keys(this.keywords);
        var cand = keys.filter(w => w.indexOf(key) == 0).sort((a, b) => a.length - b.length)
            .concat(keys.filter(w => w.indexOf(key) > 0));
        if (key == "|")
            cand.splice(1, 0, "_");
        if (key.length == 0 || cand.length == 0) {
            this.$candy.css("visibility", "hidden");
            this.candIndex = -1;
            return;
        }
        if (this.candIndex < 0) {
            var pos = this.$active.position();
            this.candIndex = 0;
            this.$candy.css({
                "visibility": "visible",
                "left": pos.left,
                "top": pos.top + this.$active.height()
            });
        }
        this.candCount = cand.length;
        var i0 = 0;
        if (cand.length > this.candMax) {
            i0 = this.candIndex - this.candMax / 2;
            if (i0 < 0)
                i0 = 0;
            else if (i0 > this.candCount - this.candMax)
                i0 = this.candCount - this.candMax;
            cand = cand.slice(i0, i0 + this.candMax);
        }
        this.$candy.empty();
        cand.forEach((c, i) => {
            var glyph = (c in this.keywords ? this.keywords[c] : c);
            var e = $("<div/>").addClass("candidate").text(c + " " + glyph);
            if (i == 0 && i0 > 0)
                e.addClass("candidateSucc");
            else if (i == cand.length - 1 && i0 + i < this.candCount - 1)
                e.addClass("candidateLast");
            else if (i0 + i == this.candIndex) {
                e.addClass("candidateSelected");
                this.candSelected = c;
            }
            this.$candy.append(e);
        });
    }
    //////////////////////////////////////
    /*  undo							*/
    //////////////////////////////////////
    undo() {
        var i = this.records.length - 1;
        var dest = this.activeField;
        while (i >= 0 && this.records[i].type == RecordType.Transfer) {
            if (this.inMacroMode && i <= this.macroOption.epoch + 2) {
                this.exitMacroMode(false);
                return;
            }
            dest = this.rollbackTransfer(dest, this.records[i]);
            i--;
        }
        this.activeField = dest;
        if (i >= 0)
            this.rollbackEdit(dest, this.records[i]);
        this.records.splice(i, this.records.length - i + 1);
    }
    rollbackTransfer(dest, rt) {
        if (rt.deeper)
            return dest.parent;
        var t;
        if ("from" in rt) {
            var rdt = rt;
            if (dest instanceof Diagram) {
                var d = dest;
                var k = d.findArrow(rdt.from, rdt.to, rdt.n);
                t = d.arrows[k.row][k.col][k.i].label;
            }
            else
                console.error("[Application.rollbackTransfer] inconsistent transfer record (arrow label)");
        }
        else
            t = dest.token(rt.index);
        if (t instanceof Structure || t instanceof Formula)
            dest = t;
        else
            console.error("[Application.rollbackTransfer] inconsistent transfer record");
        return dest;
    }
    rollbackEdit(dest, r) {
        if (r.type == RecordType.Edit) {
            var re = r;
            if (re.insert) {
                var to;
                if (re.contents.length == 1 && re.contents[0] instanceof Matrix
                    && this.activeField instanceof Matrix) {
                    var a = this.activeField;
                    var p = a.pos(re.index);
                    var m = re.contents[0];
                    to = re.index
                        + (Math.min(m.rows, a.rows - p.row) - 1) * a.cols
                        + (Math.min(m.cols, a.cols - p.col) - 1);
                }
                else
                    to = re.index + re.contents.length;
                dest.remove(re.index, to);
                this.activeIndex = Math.min(re.index, to);
            }
            else {
                this.activeIndex = dest.paste(re.index, re.contents);
            }
        }
        else if (r.type == RecordType.DiagramEdit) {
            var rea = r;
            var d;
            if (dest instanceof Diagram)
                d = dest;
            else
                console.error("[Application.rollbackEdit] incosistent record (diagram)");
            if (rea.insert)
                d.removeArrow(rea.option.from, rea.index, 0);
            else
                d.addArrow(rea.option.from, rea.index, rea.option.num, rea.option.style, rea.option.head);
        }
        else if (r.type == RecordType.DiagramDeco) {
            var rdd = r;
            var d;
            if (dest instanceof Diagram)
                d = dest;
            else
                console.error("[Application.rollbackEdit] incosistent record (diagram decolation)");
            switch (rdd.command) {
                case "f":
                    d.toggleFrame(rdd.index);
                    break;
                case "o":
                case "=":
                    d.alterFrameStyle(rdd.index, rdd.command == "o", rdd.command == "=");
                    if (rdd.prev === null)
                        d.toggleFrame(rdd.index);
                    break;
                case ".":
                case "d":
                    d.alterFrameStyle(rdd.index, false, false, rdd.prev);
                    if (rdd.prev === null)
                        d.toggleFrame(rdd.index);
                    break;
                case "+":
                case "-":
                    d.changeFrameSize(rdd.index, rdd.command == "-");
                    break;
            }
            this.activeIndex = rdd.index;
        }
        else if (r.type == RecordType.EditMatrix) {
            var rem = r;
            var m;
            if (dest instanceof Matrix)
                m = dest;
            else if (dest.parent instanceof Matrix)
                m = dest.parent;
            else
                console.error("[Application.rollbackEdit] incosistent record (matrix)");
            (rem.extend ? m.shrink : m.extend).bind(m)(rem.horizontal);
        }
        else
            console.error("[Application.rollbackEdit] unexpected record");
    }
    //////////////////////////////////////
    /*  mouse operation					*/
    //////////////////////////////////////
    dragSelect(e) {
        if (!this.dragFrom)
            return;
        var select = new Rect(Math.min(e.pageX, this.dragFrom.x), Math.min(e.pageY, this.dragFrom.y), Math.abs(e.pageX - this.dragFrom.x), Math.abs(e.pageY - this.dragFrom.y));
        this.selectByRect(select);
    }
    selectByRect(select, parent = this.formula) {
        var n = parent.count();
        var selected = [];
        for (var i = 0; i < n; i++) {
            var t = parent.token(i);
            var rect = Rect.fromJQuery(t.renderedElem);
            if ((t instanceof Structure || t instanceof Formula) && rect.contains(select))
                return this.selectByRect(select, t);
            if (select.contains(rect.center()))
                selected.push(i);
        }
        if (parent != this.activeField)
            this.jumpFormula(parent);
        this.dragRect = select;
        if (selected.length == 0) {
            this.markedIndex = this.activeIndex;
            this.render();
            return;
        }
        if (parent instanceof Matrix) {
            this.markedIndex = selected[0];
            this.activeIndex = selected[selected.length - 1];
        }
        else {
            this.markedIndex = selected[0];
            this.activeIndex = selected[selected.length - 1] + 1;
        }
        this.render();
    }
    jumpTo(p, parent = this.formula) {
        var n = parent.count();
        var distMin = Number.MAX_VALUE;
        var indexNear = -1;
        for (var i = 0; i < n; i++) {
            var t = parent.token(i);
            var rect = Rect.fromJQuery(t.renderedElem);
            if ((t instanceof Structure || t instanceof Formula) && rect.contains(p))
                return this.jumpTo(p, t);
            var g = rect.center();
            var d = normSquared(p.x - g.x, p.y - g.y);
            if (d < distMin) {
                distMin = d;
                indexNear = (p.x < g.x ? i : i + 1);
            }
        }
        if (indexNear < 0)
            return;
        if (parent != this.activeField)
            this.jumpFormula(parent);
        this.activeIndex = indexNear;
        this.render();
    }
    //////////////////////////////////////
    /*  input interpretation			*/
    //////////////////////////////////////
    interpretInput(forceTrans = false, support = true) {
        var t = null;
        var input = this.currentInput;
        if (this.inputType == InputType.Number && this.postInput == "")
            this.pushNumber();
        else if ((forceTrans
            || input.length > 1 && input != "Vert"
            || input.length == 1 && !(this.inputType == InputType.String || input in this.bracketCor))
            && (this.symbols.indexOf(input) >= 0 || input in this.keywords))
            this.pushCommand();
        else
            this.pushSymbols(support);
        this.currentInput = "";
        this.inputType = InputType.Empty;
        this.inputEscaped = false;
    }
    pushNumber() {
        var t = null;
        var input = this.currentInput;
        if (input.match("[0-9]+(\.[0-9]*)?")) {
            this.insertToken(new Num(input));
        }
    }
    pushCommand() {
        var input = this.currentInput;
        var struct;
        var style = FontStyle.Normal;
        switch (input) {
            case "infer":
            case "/":
            case "frac":
                struct = new Structure(this.activeField, input == "infer" ? StructType.Infer : StructType.Frac);
                struct.elems[0] = new Formula(struct);
                struct.elems[1] = new Formula(struct);
                if (struct.type == StructType.Infer)
                    struct.elems[2] = new Formula(struct);
                this.insertToken(struct, last => input != "frac"
                    && !(last instanceof Symbol && last.str == "&"));
                break;
            case "^":
            case "_":
                if (this.activeField.parent
                    && this.activeField.parent instanceof BigOpr
                    && this.activeIndex == 0)
                    break;
                struct = new Structure(this.activeField, input == "^" ? StructType.Power : StructType.Index);
                struct.elems[0] = new Formula(struct);
                this.insertToken(struct);
                break;
            case "matrix":
            case "pmatrix":
            case "bmatrix":
            case "Bmatrix":
            case "vmatrix":
            case "Vmatrix":
                struct = new Matrix(this.activeField, 1, 1);
                struct.elems[0] = new Formula(struct);
                var br = this.keywords[input];
                if (br != "") {
                    var f = new Formula(this.activeField, br, this.bracketCor[br]);
                    this.insertToken(f);
                    struct.parent = f;
                }
                this.insertToken(struct);
                break;
            case "(":
            case "[":
            case "{":
            case "|":
            case "Vert":
            case "floor":
            case "ceil":
            case "angle":
            case "sqrt":
                var br = this.keywords[input];
                this.insertToken(new Formula(this.activeField, br, this.bracketCor[br]));
                break;
            case "mathbf":
            case "mathrm":
            case "mathscr":
            case "mathfrak":
            case "mathbb":
            case "mathtt":
                this.insertToken(new Formula(this.activeField, "", "", styles[input]));
                break;
            case "grave":
            case "acute":
            case "hat":
            case "tilde":
            case "bar":
            case "breve":
            case "dot":
            case "ddot":
            case "mathring":
            case "check":
                this.postInput = this.keywords[input];
                break;
            case "widetilde":
            case "widehat":
            case "overleftarrow":
            case "overrightarrow":
            case "overline":
            case "underline":
            case "overbrace":
            case "underbrace":
                this.insertToken(new Accent(this.activeField, this.keywords[input], input != "underline" && input != "underbrace"));
                break;
            case "xymatrix":
                this.insertToken(new Diagram(this.activeField, 1, 1));
                break;
            default:
                if (input in this.macroTable) {
                    this.insertToken(new Macro(this.activeField, input, this.macroTable[input].argc));
                }
                else if (input in this.keywords &&
                    this.operators.indexOf(this.keywords[input]) >= 0) {
                    struct = new BigOpr(this.activeField, this.keywords[input]);
                    struct.elems[0] = new Formula(struct);
                    struct.elems[1] = new Formula(struct);
                    this.insertToken(struct);
                }
                else {
                    var s = (input in this.keywords)
                        ? new Symbol(this.keywords[input] + this.postInput, false)
                        : new Symbol(input, this.inputType == InputType.String);
                    this.insertToken(s);
                    this.postInput = "";
                }
                break;
        }
    }
    pushSymbols(support = true) {
        var t;
        var input;
        if (this.inMacroMode
            && this.currentInput.match(/^#\d+$/)) {
            this.insertToken(new Symbol(this.currentInput, false));
            this.postInput = "";
            return;
        }
        if (this.currentInput == "Vert")
            input = ["‖"];
        else {
            if (this.currentInput == "")
                return;
            input = this.currentInput.split("");
        }
        for (var i = 0; i < input.length; i++) {
            t = new Symbol(input[i] + this.postInput, this.inputType == InputType.String);
            this.insertToken(t);
        }
        if (support && t.str in this.bracketCor) {
            this.insertToken(new Symbol(this.bracketCor[t.str], false));
            this.activeIndex--;
        }
        this.postInput = "";
    }
    //////////////////////////////////////
    /*  diagram editing					*/
    //////////////////////////////////////
    decorateObject(command) {
        console.log("decorate " + this.activeIndex + " " + command);
        if (!(this.activeField instanceof Diagram))
            return false;
        var d = this.activeField;
        var prev = null;
        if (this.activeIndex in d.decorations) {
            var p = d.pos(this.activeIndex);
            if (d.decorations[p.row][p.col])
                prev = d.decorations[p.row][p.col].style;
        }
        switch (command) {
            case "f":
                d.toggleFrame(this.activeIndex);
                break;
            case "o":
                d.alterFrameStyle(this.activeIndex, true);
                break;
            case "=":
                d.alterFrameStyle(this.activeIndex, false, true);
                break;
            case "-":
                d.changeFrameSize(this.activeIndex, false);
                break;
            case "d":
                d.alterFrameStyle(this.activeIndex, false, false, (prev == StrokeStyle.Dashed ? StrokeStyle.Plain : StrokeStyle.Dashed));
                break;
            case ".":
                d.alterFrameStyle(this.activeIndex, false, false, StrokeStyle.Dotted);
                break;
            case "+":
                d.changeFrameSize(this.activeIndex, true);
                break;
            default:
                return false;
        }
        var rec = {
            type: RecordType.DiagramDeco,
            index: this.activeIndex,
            command: command,
            prev: prev
        };
        this.records.push(rec);
        return true;
    }
    addArrow() {
        console.log("arrow " + this.diagramOption.from + " -> " + this.activeIndex);
        if (!(this.activeField instanceof Diagram))
            return;
        var d = this.activeField;
        d.addArrow(this.diagramOption.from, this.activeIndex, this.diagramOption.num, this.diagramOption.style, this.diagramOption.head);
        var rec = {
            type: RecordType.DiagramEdit,
            index: this.activeIndex,
            insert: true,
            option: $.extend({}, this.diagramOption)
        };
        this.records.push(rec);
        this.activeIndex = this.diagramOption.from;
        this.diagramOption.from = -1;
    }
    removeArrow() {
        console.log("remove " + this.diagramOption.from + " -> " + this.activeIndex);
        if (!(this.activeField instanceof Diagram))
            return;
        var d = this.activeField;
        var removed = d.removeArrow(this.diagramOption.from, this.activeIndex, 0);
        var i = removed.from.row * d.cols + removed.from.col;
        var rec = {
            type: RecordType.DiagramEdit,
            index: this.activeIndex,
            insert: false,
            option: $.extend(removed, { from: i })
        };
        this.records.push(rec);
    }
    labelArrow(pos) {
        console.log("label " + this.diagramOption.from + " -> " + this.activeIndex);
        if (!(this.activeField instanceof Diagram))
            return;
        var d = this.activeField;
        var a = d.labelArrow(this.diagramOption.from, this.activeIndex, this.diagramOption.arrowIndex, pos);
        var rec = {
            type: RecordType.Transfer,
            index: this.activeIndex,
            from: this.diagramOption.from,
            to: this.diagramOption.to,
            n: this.diagramOption.arrowIndex,
            deeper: true
        };
        this.records.push(rec);
        this.diagramOption.to = this.activeIndex;
        this.activeIndex = 0;
        this.activeField = a.label;
    }
    //////////////////////////////////////
    /*  matrix editing					*/
    //////////////////////////////////////
    modifyMatrix(horizontal, extend) {
        var leave = false;
        var m;
        if (this.activeField instanceof Matrix)
            m = this.activeField;
        else if (this.activeField.parent instanceof Matrix) {
            leave = true;
            this.leaveFormula(false, true);
            m = this.activeField;
        }
        else
            return;
        if (!extend) {
            if (horizontal && m.nonEmpty(0, m.cols - 1, m.rows, 1))
                this.removeToken(m.cols - 1, m.cols * m.rows - 1, true);
            else if (!horizontal && m.nonEmpty(m.rows - 1, 0, 1, m.cols))
                this.removeToken(m.cols * (m.rows - 1), m.cols * m.rows - 1, true);
        }
        (extend ? m.extend : m.shrink).bind(m)(horizontal);
        var rec = {
            type: RecordType.EditMatrix,
            index: this.activeIndex,
            extend: extend,
            horizontal: horizontal
        };
        this.records.push(rec);
        if (leave)
            this.enterFormula(false);
    }
    //////////////////////////////////////
    /*  activeFormula transition		*/
    //////////////////////////////////////
    transferFormula(forward, target) {
        var adj;
        var p = this.activeField.parent;
        if (p !== null) {
            if (target)
                adj = target;
            else {
                var a = (forward ? p.next : p.prev)(this.activeField);
                if (!(a instanceof Formula))
                    return false;
                adj = a;
            }
            var rec1 = {
                type: RecordType.Transfer,
                index: this.activeField.parent.indexOf(this.activeField),
                deeper: false
            };
            var rec2 = {
                type: RecordType.Transfer,
                index: p.indexOf(adj),
                deeper: true
            };
            this.records.push(rec1);
            this.records.push(rec2);
            this.activeField = adj;
            this.activeIndex = (forward ? 0 : adj.count());
            return true;
        }
        return false;
    }
    leaveFormula(forward, single) {
        var t = this.activeField;
        if (this.inMacroMode && t == this.macroOption.field
            && !this.exitMacroMode(false))
            return;
        if (t.parent instanceof Structure
            && !(single
                || this.markedIndex >= 0 && t.parent instanceof Matrix
                || t.parent instanceof Diagram)) {
            var rec0 = {
                type: RecordType.Transfer,
                index: t.parent.indexOf(t),
                deeper: false
            };
            this.records.push(rec0);
            t = t.parent;
        }
        var f = t.parent;
        var inFormula = f instanceof Formula;
        var rec = {
            type: RecordType.Transfer,
            index: f.indexOf(t),
            deeper: false
        };
        if (f instanceof Diagram && this.diagramOption.from >= 0) {
            rec.from = this.diagramOption.from;
            rec.to = this.diagramOption.to;
            rec.n = this.diagramOption.arrowIndex;
            this.activeIndex = forward ? this.diagramOption.to : this.diagramOption.from;
            this.diagramOption.from = -1;
            this.diagramOption.to = -1;
            this.diagramOption.arrowIndex = -1;
        }
        else {
            this.activeIndex = rec.index;
            if (inFormula && forward)
                this.activeIndex++;
        }
        if (this.markedIndex >= 0) {
            this.markedIndex = rec.index;
            if (inFormula && !forward)
                this.markedIndex++;
        }
        this.records.push(rec);
        this.activeField = f;
        return true;
    }
    enterFormula(forward) {
        var i = this.activeIndex;
        if (this.activeField instanceof Formula && !forward)
            i--;
        var dest = this.activeField.token(i);
        if (this.markedIndex < 0 && dest
            && (dest instanceof Structure && dest.count() > 0
                || dest instanceof Formula)) {
            var rec = {
                type: RecordType.Transfer,
                index: i,
                deeper: true
            };
            this.records.push(rec);
            if (dest instanceof Structure && !(dest instanceof Diagram)) {
                var s = dest;
                var j = forward ? 0 : s.elems.length - 1;
                this.activeField = s.token(j);
                var rec2 = {
                    type: RecordType.Transfer,
                    index: j,
                    deeper: true
                };
                this.records.push(rec2);
                if (dest instanceof Diagram)
                    this.diagramOption.from = -1;
            }
            else
                this.activeField = dest;
            if (forward)
                this.activeIndex = 0;
            else {
                this.activeIndex = this.activeField.count();
                if (this.activeIndex > 0 && dest instanceof Matrix)
                    this.activeIndex--;
            }
            return true;
        }
        return false;
    }
    jumpFormula(target) {
        var leave = 0, enter = -1;
        var toSeq = [];
        var toIndex = [];
        for (var to = target; to; to = to.parent) {
            toSeq.push(to);
            if (to.parent)
                toIndex.push(to.parent.indexOf(to));
        }
        outer: for (var from = this.activeField; from; from = from.parent) {
            for (var j = 0; j < toSeq.length; j++)
                if (from == toSeq[j]) {
                    enter = j;
                    break outer;
                }
            leave++;
        }
        if (enter < 0)
            console.error("[Application.jumpFormula] ill-structured formula");
        var t = this.activeField;
        for (var i = 0; i < leave; i++) {
            var f = t.parent;
            var rec = {
                type: RecordType.Transfer,
                index: f.indexOf(t),
                deeper: false
            };
            this.records.push(rec);
            t = t.parent;
        }
        for (var j = enter - 1; j >= 0; j--) {
            var rec = {
                type: RecordType.Transfer,
                index: toIndex[j],
                deeper: true
            };
            this.records.push(rec);
        }
        this.activeField = target;
        return true;
    }
    //////////////////////////////////////
    /*  activeFormula editing			*/
    //////////////////////////////////////
    insertToken(t, capture) {
        console.log("insert " + t.toString() + " at " + this.activeIndex + (capture ? " with capture" : ""));
        if (!(this.activeField instanceof Formula))
            return;
        var f = this.activeField;
        var captured = false;
        if (t instanceof Structure) {
            var struct = t;
            var last;
            if (this.activeIndex > 0 && capture
                && capture(last = f.tokens[this.activeIndex - 1])) {
                captured = true;
                struct.elems[0].insert(0, last);
                this.removeToken(this.activeIndex - 1, this.activeIndex);
            }
        }
        f.insert(this.activeIndex, t);
        var rec = {
            type: RecordType.Edit,
            index: this.activeIndex,
            insert: true,
            contents: [t.clone(null)]
        };
        this.records.push(rec);
        if (t instanceof Structure && t.count() > 0
            || t instanceof Formula) {
            this.enterFormula(true);
            if (captured)
                this.transferFormula(true);
        }
        else
            this.activeIndex++;
    }
    // "paste" method rewrites token's parent
    pasteToken(t) {
        if (this.activeField instanceof Matrix
            && t.length == 1 && t[0] instanceof Matrix) {
            var a = this.activeField;
            var p = a.pos(this.activeIndex);
            var m = t[0];
            var mr = Math.min(m.rows, a.rows - p.row);
            var mc = Math.min(m.cols, a.cols - p.col);
            if (a.nonEmpty(p.row, p.col, mr, mc))
                this.removeToken(this.activeIndex, this.activeIndex + (mr - 1) * a.cols + (mc - 1));
        }
        console.log("paste " + t.toString());
        var rec = {
            type: RecordType.Edit,
            index: this.activeIndex,
            insert: true,
            contents: t.map(x => x.clone(null))
        };
        this.records.push(rec);
        this.activeIndex = this.activeField.paste(this.activeIndex, t);
    }
    removeToken(from, to, extensive) {
        console.log("remove " + from + " ~ " + to);
        var removed;
        if (extensive && this.activeField instanceof Diagram)
            removed = this.activeField.remove(from, to, true);
        else
            removed = this.activeField.remove(from, to);
        var index = Math.min(from, to);
        var rec = {
            type: RecordType.Edit,
            index: index,
            insert: false,
            contents: removed
        };
        this.records.push(rec);
        this.activeIndex = index;
    }
    extrudeToken(from, to) {
        console.log("extrude " + from + " ~ " + to);
        var target = this.activeField.remove(from, to);
        var index = Math.min(from, to);
        var extruded = [new Formula(null)];
        extruded[0].tokens = target;
        var rec1 = {
            type: RecordType.Edit,
            index: index,
            insert: false,
            contents: target
        };
        var rec2 = {
            type: RecordType.Edit,
            index: index,
            insert: true,
            contents: extruded
        };
        this.records.push(rec1);
        this.records.push(rec2);
        this.activeField.paste(index, extruded);
        this.activeIndex = index;
    }
    //////////////////////////////////////
    /*  formula output					*/
    //////////////////////////////////////
    outputToken(q, t, info) {
        var e;
        if (t instanceof Symbol) {
            e = this.outputSymbol(q, t, info);
        }
        else if (t instanceof Num) {
            e = $("<div/>")
                .addClass("number")
                .text(t.value.toString());
            q.append(e);
        }
        else if (t instanceof Macro) {
            var m = t;
            e = this.outputToken(q, this.macroTable[m.name].content, {
                arguments: m, actGenerated: info.actGenerated
            });
        }
        else if (t instanceof Structure) {
            var s = t;
            e = this.outputStruct(s, info);
            q.append(e);
            if (s.type == StructType.Infer) {
                var a3 = $("<div/>").addClass("math label");
                this.outputToken(a3, s.token(2), info);
                q.append(a3);
            }
        }
        else if (t instanceof Formula) {
            e = this.outputFormula(t, info);
            q.append(e);
        }
        else
            console.error("[Application.outputToken] unexpected argument : " + t);
        t.renderedElem = e;
        return e;
    }
    outputSymbol(q, s, info) {
        var str = s.str;
        if (info.arguments) {
            var m = str.match(/^#(\d+)$/);
            if (m) {
                var arg = info.arguments;
                info.arguments = null;
                var e = this.outputToken(q, arg.token(parseInt(m[1]) - 1), info);
                info.arguments = arg;
                return e;
            }
        }
        var style = this.outputCurrentStyle[0];
        if (style != FontStyle.Normal)
            str = this.transStyle(str, style);
        if (str == "&")
            str = EmSpace;
        var e = $("<div/>").text(str);
        if (style != FontStyle.Normal)
            e.addClass("styledLetter");
        if (!this.proofMode && s.variable && (style == FontStyle.Normal || style == FontStyle.Bold))
            e.addClass("variable");
        else
            e.addClass("symbol");
        q.append(e);
        return e;
    }
    transStyle(str, style) {
        var table;
        switch (style) {
            case FontStyle.Bold:
                table = Bold;
                break;
            case FontStyle.Script:
                table = Script;
                break;
            case FontStyle.Fraktur:
                table = Fraktur;
                break;
            case FontStyle.BlackBoard:
                table = DoubleStruck;
                break;
            case FontStyle.Roman:
                table = SansSerif;
                break;
            case FontStyle.Typewriter:
                table = Monospace;
                break;
            default: console.error("[Application.transStyle] unexpected font style : " + style);
        }
        var r = "";
        for (var i = 0; i < str.length; i++) {
            var c = str.charAt(i);
            r += c in table ? table[c] : c;
        }
        return r;
    }
    outputStruct(s, info) {
        var e;
        switch (s.type) {
            case StructType.Frac:
            case StructType.Infer:
                e = $("<div/>").addClass("frac");
                var prim = this.outputToken(e, s.token(0), info);
                var seco = this.outputToken(e, s.token(1), info);
                if (s.type == StructType.Infer) {
                    e.addClass("reverseOrdered");
                    prim.addClass("overline");
                }
                else
                    seco.addClass("overline");
                break;
            case StructType.Power:
            case StructType.Index:
                e = $("<div/>").addClass(s.type == StructType.Power ? "power" : "index");
                this.outputToken(e, s.token(0), info);
                break;
            case StructType.Diagram:
                this.afterLayout.push(s);
            case StructType.Matrix:
                var m = s;
                e = $("<div/>").addClass("matrix");
                if (this.activeField == m && this.markedIndex >= 0) {
                    var mark = true;
                    var ai = Math.floor(this.activeIndex / m.cols);
                    var aj = this.activeIndex % m.cols;
                    var mi = Math.floor(this.markedIndex / m.cols);
                    var mj = this.markedIndex % m.cols;
                    var i1 = Math.min(ai, mi);
                    var j1 = Math.min(aj, mj);
                    var i2 = Math.max(ai, mi);
                    var j2 = Math.max(aj, mj);
                }
                for (var i = 0; i < m.rows; i++) {
                    var r = $("<div/>").addClass("row");
                    for (var j = 0; j < m.cols; j++) {
                        var c = $("<div/>").addClass(m.type == StructType.Diagram ? "xycell" : "cell");
                        var t = this.outputToken(c, m.tokenAt(i, j), info);
                        if (m == this.activeField) {
                            var k = i * m.cols + j;
                            if (k == this.diagramOption.from)
                                t.addClass("arrowStart");
                            if (k == this.activeIndex)
                                t.addClass("active");
                        }
                        if (mark && i >= i1 && i <= i2 && j >= j1 && j <= j2)
                            c.addClass("marked");
                        r.append(c);
                    }
                    e.append(r);
                }
                break;
            case StructType.BigOpr:
                var o = s;
                if (["∫", "∮", "∬", "∭", "⨌"].indexOf(o.operator) >= 0) {
                    e = $("<div/>").addClass("math");
                    e.append($("<div/>").text(o.operator).addClass("operator"));
                    var f = $("<div/>").addClass("frac");
                    this.outputToken(f, s.token(1), info).addClass("subFormula");
                    this.outputToken(f, s.token(0), info).addClass("subFormula");
                    e.append(f);
                }
                else {
                    e = $("<div/>").addClass("frac");
                    this.outputToken(e, s.token(1), info).addClass("subFormula");
                    e.append($("<div/>").text(o.operator).addClass("operator"));
                    this.outputToken(e, s.token(0), info).addClass("subFormula");
                }
                break;
            case StructType.Accent:
                var a = s;
                if (a.symbol == "‾")
                    e = this.outputFormula(a.elems[0], info).addClass("overline");
                else if (a.symbol == "_")
                    e = this.outputFormula(a.elems[0], info).addClass("underline");
                else {
                    e = $("<div/>").addClass("frac");
                    if (!a.above)
                        e.addClass("reverseOrdered");
                    var ac = this.makeGlyph(a.symbol).addClass("accent").text(EnSpace);
                    e.append(ac);
                    this.outputToken(e, s.token(0), info);
                }
                break;
        }
        return e;
    }
    outputFormula(f, info) {
        var r;
        var shift = false;
        if (f.style != this.outputCurrentStyle[0]) {
            this.outputCurrentStyle.unshift(f.style);
            shift = true;
        }
        if (f.prefix != "" || f.suffix != "") {
            var braced = $("<div/>").addClass("embraced");
            if (f.prefix != "")
                braced.append(this.makeGlyph(f.prefix).addClass("bracket"));
            var inner = this.outputFormulaInner(f, info);
            if (f.prefix == "√")
                inner.addClass("overline");
            braced.append(inner);
            if (f.suffix != "")
                braced.append(this.makeGlyph(f.suffix).addClass("bracket"));
            r = braced;
        }
        else
            r = this.outputFormulaInner(f, info);
        if (shift) {
            r.addClass("formulaStyled");
            this.outputCurrentStyle.shift();
        }
        return r;
    }
    makeGlyph(char) {
        var q = $("<div/>");
        var dat = this._glyph.generate(char);
        if (dat != "")
            q = q.css("background-image", "url(" + dat + ")");
        else
            q = q.text(char);
        return q;
    }
    outputFormulaInner(f, info) {
        var e = $("<div/>").addClass(this.proofMode ? "formula" : "math");
        if (f == this.macroOption.field)
            e.addClass("macroField");
        if (f == this.activeField && !info.actGenerated) {
            var r;
            var markedFrom = Math.min(this.markedIndex, this.activeIndex);
            var markedTo = Math.max(this.markedIndex, this.activeIndex);
            var marked = false;
            for (var i = 0, j = 0; i <= f.count(); i++) {
                if (i == this.activeIndex) {
                    this.$active = $("<div/>");
                    if (this.markedIndex < 0)
                        this.$active.addClass("active");
                    e.append(this.$active);
                }
                if (this.markedIndex >= 0) {
                    if (j == markedFrom) {
                        r = $("<div/>").addClass("math marked");
                        e.append(r);
                        marked = true;
                    }
                    if (j == markedTo)
                        marked = false;
                }
                if (j == f.count())
                    break;
                this.outputToken(marked ? r : e, f.tokens[j++], info);
            }
            info.actGenerated = true;
        }
        else if (f.tokens.length > 0)
            f.tokens.forEach(s => {
                this.outputToken(e, s, info);
            });
        else
            e.append($("<div/>").addClass("blank").text(EnSpace));
        return e;
    }
    drawAfterLayout() {
        var box = this.$field[0].getBoundingClientRect();
        // var box = this.ghost[0].getBoundingClientRect();
        this.$ghost.prop({
            "width": box.width,
            "height": box.height
        });
        var ctx = this.$ghost[0].getContext("2d");
        for (var i = 0; i < this.afterLayout.length; i++) {
            if (this.afterLayout[i] instanceof Diagram)
                this.drawDiagram(ctx, box, this.afterLayout[i]);
        }
    }
    drawDiagram(ctx, box, d) {
        d.decorations.forEach((r, i) => r.forEach((deco, j) => {
            if (deco)
                d.drawFrame(ctx, box, i * d.cols + j, deco);
        }));
        this.drawArrows(ctx, box, d);
    }
    drawArrows(ctx, box, d) {
        var selected = false;
        var from = d.pos(this.diagramOption.from);
        var to = d.pos(this.activeIndex);
        d.arrows.forEach((ar, i) => ar.forEach((ac, j) => groupBy(ac, a => a.to.row * d.cols + a.to.col)
            .forEach(as => {
            var active = (i == from.row && j == from.col
                && as[0].to.row == to.row && as[0].to.col == to.col);
            if (active) {
                if (this.diagramOption.arrowIndex < 0 || this.diagramOption.arrowIndex >= as.length)
                    this.diagramOption.arrowIndex = 0;
            }
            as.forEach((a, k) => {
                var label = null;
                if (!a.label.empty() || this.activeField == a.label) {
                    var label = $("<div/>").addClass("arrowLabel");
                    // this implementation unable to macroize diagram
                    this.outputToken(label, a.label, { arguments: null, actGenerated: false });
                    if (this.activeField == a.label && a.label.empty())
                        this.$active.text(EnSpace); // there must be some contents to layout in drawArrow
                    this.$field.append(label);
                }
                var shift = 10 * (k - (as.length - 1) / 2);
                if (active && k == this.diagramOption.arrowIndex) {
                    var color = $("<div/>").addClass("activeArrow").css("color");
                    d.drawArrow(ctx, box, label, a, shift, this.activeField == d ? color : null);
                    selected = true;
                }
                else
                    d.drawArrow(ctx, box, label, a, shift);
            });
        })));
        if (d == this.activeField && this.diagramOption.from >= 0 && !selected) {
            var a = {
                from: d.pos(this.diagramOption.from),
                to: d.pos(this.activeIndex),
                style: this.diagramOption.style,
                head: this.diagramOption.head,
                num: this.diagramOption.num,
                label: null, labelPos: null
            };
            var color = $("<div/>").addClass("intendedArrow").css("color");
            d.drawArrow(ctx, box, null, a, 0, color);
        }
        if (!selected && d == this.activeField)
            this.diagramOption.arrowIndex = -1;
    }
}

export default TypeMath;