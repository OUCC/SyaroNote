(function(global) {
  "use strict;"

  // Class
  function TableFormatter() {
    HashHandler = ace.require('ace/keyboard/hash_handler').HashHandler;

    return new HashHandler([{
      bindKey: 'Tab',
      descr: "Format markdown table",
      exec: function(ed) {
        var row = ed.getCursorPosition().row;
        var currLine = ed.getSession().getLine(row);
        console.debug(currLine);
        if(currLine.slice(0, 1) == '|') {
            return true; // avoid insert \t or 4 space
          } else {
            return false; // allow other ace commands to handle event
          }
        }
      }])
  };

  // export
  global['TableFormatter'] = TableFormatter;

})((this || 0).self || global);
