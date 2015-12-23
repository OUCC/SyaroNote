export function get (wpath) {
  var url = '/api/get?wpath=' + encodeURIComponent(wpath);

  return xhr({
    url: url,
    method: 'GET',
  });
}

export function update (wpath, contents, message, name, email) {
  var url = '/api/update?wpath=' + encodeURIComponent(wpath);
  if (message) { url += '&message=' + encodeURIComponent(message); }
  if (name) { url += '&name=' + encodeURIComponent(name); }
  if (email) { url += '&email=' + encodeURIComponent(email); }

  return xhr({
    url: url,
    method: 'POST',
    body: contents,
  });
}

function xhr(arg) {
  return new Promise(function (resolve, reject) {
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
      if (Math.floor(req.status / 100) === 2) {
        resolve(arg);
      } else {
        reject(arg);
      }
    };

    req.onerror = function () {
      reject(arg);
    };

    if (arg.body) req.send(arg.body);
    else req.send();
  });
}
