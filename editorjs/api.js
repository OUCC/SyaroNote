export function get (wpath, callback) {
  var url = '/api/get?wpath=' + encodeURIComponent(wpath);

  var xhr = new XMLHttpRequest();
  xhr.open('GET', url);
  xhr.onreadystatechange = function () {
    if(xhr.readyState !== 4) { return; }
    switch (xhr.status) {
    case 200:
      callback(xhr.responseText);
      break

    default:
      callback(undefined, xhr.statusText);
      break
    }
  };
  xhr.send();
}

export function update (wpath, contents, callback, message, name, email) {
  var url = '/api/update?wpath=' + encodeURIComponent(wpath);
  if (message) { url += '&message=' + encodeURIComponent(message); }
  if (name) { url += '&name=' + encodeURIComponent(name); }
  if (email) { url += '&email=' + encodeURIComponent(email); }

  var xhr = new XMLHttpRequest();
  xhr.open('POST', url);
  xhr.onreadystatechange = function () {
    if(xhr.readyState !== 4) { return; }
    switch (xhr.status) {
    case 200:
      callback();
      break

    default:
      callback(xhr.statusText);
      break
    }
  };
  xhr.send(contents);
}
