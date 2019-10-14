import Toc from "./Toc.svelte";
import SearchInput from "./SearchInput.svelte";
import { item } from "./item.js";

// polyfil for Object.is
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Object/is
if (!Object.is) {
  Object.is = function (x, y) {
    // SameValue algorithm
    if (x === y) {
      // Steps 1-5, 7-10
      // Steps 6.b-6.e: +0 != -0
      return x !== 0 || 1 / x === 1 / y;
    } else {
      // Step 6.a: NaN == NaN
      return x !== x && y !== y;
    }
  };
}

// pageId looks like "5ab3b56329c44058b5b24d3f364183ce"
// find full url of the page matching this pageId
function findURLWithPageId(pageId) {
  var n = gBookToc.length;
  for (var i = 0; i < n; i++) {
    var tocItem = gBookToc[i];
    var uri = item.url(tocItem);
    // uri looks like "go-get-5ab3b56329c44058b5b24d3f364183ce"
    if (uri.endsWith(pageId)) {
      return uri;
    }
  }
  return "";
}

function do404() {
  var loc = window.location.pathname;
  var locParts = loc.split("/");
  var lastIdx = locParts.length - 1;
  var uri = locParts[lastIdx];
  // redirect ${garbage}-${id} => ${correct url}-${id}
  var parts = uri.split("-");
  var pageId = parts[parts.length - 1];
  var fullURL = findURLWithPageId(pageId);
  if (fullURL != "") {
    locParts[lastIdx] = fullURL;
    var loc = locParts.join("/");
    window.location.pathname = loc;
  }
}

function httpsMaybeRedirect() {
  if (window.location.protocol !== "http:") {
    return;
  }
  if (window.location.hostname !== "www.programming-books.io") {
    return;
  }
  var uri = window.location.toString();
  uri = uri.replace("http://", "https://");
  window.location = uri;
}

function showcontact() {
  var el = document.getElementById("contact-form");
  el.style.display = "block";
  el = document.getElementById("contact-page-url");
  var uri = window.location.href;
  //uri = uri.replace("#", "");
  el.value = uri;
  el = document.getElementById("msg-for-chris");
  el.focus();
}

function hidecontact() {
  var el = document.getElementById("contact-form");
  el.style.display = "none";
}

window.showcontact = showcontact;
window.hidecontact = hidecontact;

const app = {
  toc: Toc,
  searchInput: SearchInput,
  do404: do404,
  httpsMaybeRedirect: httpsMaybeRedirect,
};

export default app;
