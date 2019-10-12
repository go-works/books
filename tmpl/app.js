// we're applying react-like state => UI
var currentState = {
  searchResults: [],
  // index within searchResults array, -1 means not selected
  selectedSearchResultIdx: -1
};

var currentSearchTerm = "";

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

function storeSet(key, val) {
  if (window.localStorage) {
    window.localStorage.setItem(key, val);
  }
}

function storeClear(key) {
  if (window.localStorage) {
    window.localStorage.removeItem(key);
  }
}

function storeGet(key) {
  if (window.localStorage) {
    return window.localStorage.getItem(key);
  }
  return "";
}

var keyScrollPos = "scrollPos";
var keyIndexView = "indexView";

function viewSet(view) {
  storeSet(keyIndexView, view);
}

function viewGet() {
  return storeGet(keyIndexView);
}

function viewClear() {
  storeClear(keyIndexView);
}

// rv = rememberView but short because it's part of url
function rv(view) {
  //console.log("rv:", view);
  viewSet(view);
}

// accessor functions for items in gBookToc array:
// 	[${chapter or aticle url}, ${parentIdx}, ${title}, ${synonim 1}, ${synonim 2}, ...],
// as generated in gen_book_toc_search.go and stored in books/${book}/toc_search.js

var itemIdxIsExpanded = 0;
var itemIdxURL = 1;
var itemIdxParent = 2;
var itemIdxFirstChild = 3;
var itemIdxTitle = 4;
var itemIdxFirstSynonym = 5;

function tocItemIsExpanded(item) {
  return item[itemIdxIsExpanded];
}

function tocItemSetIsExpanded(item, isExpanded) {
  item[itemIdxIsExpanded] = isExpanded;
}

function tocItemURL(item) {
  while (item) {
    var uri = item[itemIdxURL];
    if (uri != "") {
      return uri;
    }
    item = tocItemParent(item);
  }
  return "";
}

function tocItemFirstChildIdx(item) {
  return item[itemIdxFirstChild];
}

function tocItemHasChildren(item) {
  return tocItemFirstChildIdx(item) != -1;
}

// returns true if has children and some of them articles
// (as opposed to children that are headers within articles)
function tocItemHasArticleChildren(item) {
  var idx = tocItemFirstChildIdx(item);
  if (idx == -1) {
    return false;
  }
  var item = gBookToc[idx];
  var parentIdx = item[itemIdxParent];
  while (idx < gBookToc.length) {
    item = gBookToc[idx];
    if (parentIdx != item[itemIdxParent]) {
      return false;
    }
    var uri = item[itemIdxURL];
    if (uri.indexOf("#") === -1) {
      return true;
    }
    idx += 1;
  }
  return false;
}

function tocItemParent(item) {
  var idx = tocItemParentIdx(item);
  if (idx == -1) {
    return null;
  }
  return gBookToc[idx];
}

function tocItemIsRoot(item) {
  return tocItemParentIdx(item) == -1;
}

function tocItemParentIdx(item) {
  return item[itemIdxParent];
}

function tocItemTitle(item) {
  return item[itemIdxTitle];
}

// all searchable items: title + search synonyms
function tocItemSearchable(item) {
  return item.slice(itemIdxTitle);
}

// from https://github.com/component/escape-html/blob/master/index.js
var matchHtmlRegExp = /["'&<>]/;
function escapeHTML(string) {
  var str = "" + string;
  var match = matchHtmlRegExp.exec(str);

  if (!match) {
    return str;
  }

  var escape;
  var html = "";
  var index = 0;
  var lastIndex = 0;

  for (index = match.index; index < str.length; index++) {
    switch (str.charCodeAt(index)) {
      case 34: // "
        escape = "&quot;";
        break;
      case 38: // &
        escape = "&";
        break;
      case 39: // '
        escape = "&#39;";
        break;
      case 60: // <
        escape = "&lt;";
        break;
      case 62: // >
        escape = "&gt;";
        break;
      default:
        continue;
    }

    if (lastIndex !== index) {
      html += str.substring(lastIndex, index);
    }

    lastIndex = index + 1;
    html += escape;
  }

  return lastIndex !== index ? html + str.substring(lastIndex, index) : html;
}

// splits a string in two parts at a given index
// ("foobar", 2) => ["fo", "obar"]
function splitStringAt(s, idx) {
  var res = ["", ""];
  if (idx == 0) {
    res[1] = s;
  } else {
    res[0] = s.substring(0, idx);
    res[1] = s.substring(idx);
  }
  return res;
}

function tagOpen(name, opt) {
  opt = opt || {};
  var classes = opt.classes || [];
  if (opt.cls) {
    classes.push(opt.cls);
  }
  var cls = classes.join(" ");

  var s = "<" + name;
  var attrs = [];
  if (cls) {
    attrs.push(attr("class", cls));
  }
  if (opt.id) {
    attrs.push(attr("id", opt.id));
  }
  if (opt.title) {
    attrs.push(attr("title", opt.title));
  }
  if (opt.href) {
    attrs.push(attr("href", opt.href));
  }
  if (opt.onclick) {
    attrs.push(attr("onclick", opt.onclick));
  }
  if (attrs.length > 0) {
    s += " " + attrs.join(" ");
  }
  return s + ">";
}

function tagClose(tagName) {
  return "</" + tagName + ">";
}

function inTag(tagName, contentHTML, opt) {
  return tagOpen(tagName, opt) + contentHTML + tagClose(tagName);
}

function inTagRaw(tagName, content, opt) {
  var contentHTML = escapeHTML(content);
  return tagOpen(tagName, opt) + contentHTML + tagClose(tagName);
}

function attr(name, val) {
  val = val.replace("'", "");
  return name + "='" + val + "'";
}

function span(s, opt) {
  return inTagRaw("span", s, opt);
}

function div(html, opt) {
  return inTag("div", html, opt);
}

function a(uri, txt, opt) {
  txt = escapeHTML(txt);
  opt.href = uri;
  opt.title = txt.replace('"', "");
  return inTag("a", txt, opt);
}

function setState(newState, now = false) {
}

function isChapterOrArticleURL(s) {
  var isChapterOrArticle = s.indexOf("#") === -1;
  return isChapterOrArticle;
}

function navigateToSearchResult(idx) {
  var loc = window.location.pathname;
  var parts = loc.split("/");
  var lastIdx = parts.length - 1;
  var lastURL = parts[lastIdx];
  var selected = currentState.searchResults[idx];
  var tocItem = selected.tocItem;

  // either replace chapter/article url or append to book url
  var uri = tocItemURL(tocItem);
  if (isChapterOrArticleURL(lastURL)) {
    parts[lastIdx] = uri;
  } else {
    parts.push(uri);
  }
  loc = parts.join("/");
  window.location = loc;
}

// create HTML to highlight part of s starting at idx and with length len
function hilightSearchResult(txt, matches) {
  var prevIdx = 0;
  var n = matches.length;
  var res = "";
  var s = "";
  // alternate non-higlighted and highlihted strings
  for (var i = 0; i < n; i++) {
    var el = matches[i];
    var idx = el[0];
    var len = el[1];

    var nonHilightLen = idx - prevIdx;
    if (nonHilightLen > 0) {
      s = txt.substring(prevIdx, prevIdx + nonHilightLen);
      res += span(s);
    }
    s = txt.substring(idx, idx + len);
    res += span(s, { cls: "hili" });
    prevIdx = idx + len;
  }
  var txtLen = txt.length;
  nonHilightLen = txtLen - prevIdx;
  if (nonHilightLen > 0) {
    s = txt.substring(prevIdx, prevIdx + nonHilightLen);
    res += span(s);
  }
  return res;
}

// return true if term is a search synonym inside tocItem
function isMatchSynonym(tocItem, term) {
  term = term.toLowerCase();
  var title = tocItemTitle(tocItem).toLowerCase();
  return title != term;
}

function getParentTitle(tocItem) {
  var res = "";
  var parent = tocItemParent(tocItem);
  while (parent) {
    var s = tocItemTitle(parent);
    if (res) {
      s = s + " / ";
    }
    res = s + res;
    parent = tocItemParent(parent);
  }
  return res;
}

// if search matched synonym returns "${chapterTitle} / ${articleTitle}"
// otherwise empty string
function getArticlePath(tocItem, term) {
  if (!isMatchSynonym(tocItem, term)) {
    return null;
  }
  var title = tocItemTitle(tocItem);
  var parentTitle = getParentTitle(tocItem);
  if (parentTitle == "") {
    return title;
  }
  return parentTitle + " / " + title;
}



// https://github.com/Treora/scroll-into-view/blob/master/polyfill.js
// TODO: passing options = { center: true } doesn't work
function scrollElementIntoView(el, options) {
  // Use traditional scrollIntoView when traditional argument is given.
  if (options === undefined || options === true || options === false) {
    el.scrollIntoView(el, arguments);
    return;
  }

  var win = el.ownerDocument.defaultView;

  // Read options.
  if (options === undefined) options = {};
  if (options.center === true) {
    options.vertical = 0.5;
    options.horizontal = 0.5;
  } else {
    if (options.block === "start") options.vertical = 0.0;
    else if (options.block === "end") options.vertical = 0.0;
    else if (options.vertical === undefined) options.vertical = 0.0;

    if (options.horizontal === undefined) options.horizontal = 0.0;
  }

  // Fetch positional information.
  var rect = el.getBoundingClientRect();

  // Determine location to scroll to.
  var targetY =
    win.scrollY +
    rect.top -
    (win.innerHeight - el.offsetHeight) * options.vertical;
  var targetX =
    win.scrollX +
    rect.left -
    (win.innerWidth - el.offsetWidth) * options.horizontal;

  // Scroll.
  win.scroll(targetX, targetY);

  // If window is inside a frame, center that frame in the parent window. Recursively.
  if (win.parent !== win) {
    // We are inside a scrollable element.
    var frame = win.frameElement;
    scrollIntoView.call(frame, options);
  }
}


// el is [idx, len]
// sort by idx.
// if idx is the same, sort by reverse len
// (i.e. bigger len is first)
function sortSearchByIdx(el1, el2) {
  var res = el1[0] - el2[0];
  if (res == 0) {
    res = el2[1] - el1[1];
  }
  return res;
}

// returns a debouncer function. Usage:
// var debouncer = makeDebouncer(250);
// function fn() { ... }
// debouncer(fn)
function makeDebouncer(timeInMs) {
  let interval;
  return function (f) {
    clearTimeout(interval);
    interval = setTimeout(() => {
      interval = null;
      f();
    }, timeInMs);
  };
}

// TODO: maybe just use debouncer from https://gist.github.com/nmsdvid/8807205
// and do "add"EventListener("input", debounce(onSearchInputChanged, 250, false))
var searchInputDebouncer = makeDebouncer(250);

function extractIntID(id) {
  var parts = id.split("-");
  var nStr = parts[parts.length - 1];
  var n = parseInt(nStr, 10);
  return isNaN(n) ? -1 : n;
}

function getIdxFromSearchResultElementId(id) {
  if (!id) {
    return -1;
  }
  if (!id.startsWith("search-result-no-")) {
    return -1;
  }
  return extractIntID(id);
}


// If we clicked on search result list, navigate to that result.
function trySearchResultNavigate(el) {
  // Since a child element might be clicked, we need to traverse up until
  // we find desired parent or top of document.
  while (el) {
    var idx = getIdxFromSearchResultElementId(el.id);
    if (idx >= 0) {
      navigateToSearchResult(idx);
      return true;
    }
    el = el.parentNode;
  }
  return false;
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

// have to do navigation in onMouseDown because when done in onClick,
// the blur event from input element swallows following onclick, so
// I had to click twice on search result
function onMouseDown(ev) {
  var el = ev.target;
  //console.log("onMouseDown ev:", ev, "el:", el);
  if (trySearchResultNavigate(el)) {
    return;
  }
}

function onClick(ev) {
  var el = ev.target;
  //console.log("onClick ev:", ev, "el:", el);
  if (el.id === "blur-overlay") {
    dismissSearch();
    return;
  }

  setState({
    selectedSearchResultIdx: -1
  });
}

function dismissSearch() {
  setState(
    {
      selectedSearchResultIdx: -1,
    },
    true
  );
}

// when we're over elements with id "search-result-no-${id}", set this one
// as selected element
function onMouseMove(ev) {
  var el = ev.target;
  var idx = getIdxFromSearchResultElementId(el.id);
  if (idx < 0) {
    return;
  }
  //console.log("ev.target:", el, "id:", el.id, "idx:", idx);
  setState({
    selectedSearchResultIdx: idx
  });
  ev.stopPropagation();
}

function onEnter(ev) {
  var selIdx = currentState.selectedSearchResultIdx;
  if (selIdx == -1) {
    return;
  }
  navigateToSearchResult(selIdx);
}

function onEscape(ev) {
  dismissSearch();
  ev.preventDefault();
}

function onUpDown(ev) {
  // "Down" is Edge, "ArrowUp" is Chrome
  var dir = ev.key == "ArrowUp" || ev.key == "Up" ? -1 : 1;
  var results = currentState.searchResults;
  var n = results.length;
  var selIdx = currentState.selectedSearchResultIdx;
  if (n <= 0 || selIdx < 0) {
    return;
  }
  var newIdx = selIdx + dir;
  if (newIdx >= 0 && newIdx < n) {
    setState({
      selectedSearchResultIdx: newIdx
    });
    ev.preventDefault();
  }
}

function onKeyDown(ev) {
  // console.log(ev);

  if (ev.key == "Enter") {
    onEnter(ev);
    return;
  }

  if (
    ev.key == "ArrowUp" ||
    ev.key == "ArrowDown" ||
    ev.key == "Up" ||
    ev.key == "Down"
  ) {
    onUpDown(ev);
    return;
  }
}

function onSearchInputChanged(ev) {
  var s = ev.target.value;
  var fn = doSearch.bind(this, s);
  searchInputDebouncer(fn);
}

function start() {
  //console.log("started");

  document.addEventListener("keydown", onKeyDown);
  document.addEventListener("mousemove", onMouseMove);
  document.addEventListener("mousedown", onMouseDown);
  document.addEventListener("click", onClick);

  // if this is chapter or article, we generate toc
  /*
  var scrollTop = scrollPosGet() || -1;
  if (scrollTop >= 0) {
    //console.log("scrollTop:", scrollTop);
    var el = document.getElementById("toc");
    el.scrollTop = scrollTop;
    scrollPosClear();
    return;
  }
  */
  /*
  function makeTocVisible() {
    var el = document.getElementById(tocItemElementID);
    if (el) {
      scrollElementIntoView(el, true);
    } else {
      console.log(
        "tried to scroll toc item to non-existent element with id: '" +
        tocItemElementID +
        "'"
      );
    }
  }
  window.requestAnimationFrame(makeTocVisible);
  */
}

// pageId looks like "5ab3b56329c44058b5b24d3f364183ce"
// find full url of the page matching this pageId
function findURLWithPageId(pageId) {
  var n = gBookToc.length;
  for (var i = 0; i < n; i++) {
    var tocItem = gBookToc[i];
    var uri = tocItemURL(tocItem);
    // uri looks like "go-get-5ab3b56329c44058b5b24d3f364183ce"
    if (uri.endsWith(pageId)) {
      return uri;
    }
  }
  return "";
}

function updateLinkHome() {
  var view = viewGet();
  if (!view) {
    return;
  }
  var uri = "/";
  if (view === "list") {
    // do nothing
  } else if (view == "grid") {
    uri = "/index-grid";
  } else {
    console.log("unknown view:", view);
    viewClear();
  }
  var el = document.getElementById("link-home");
  if (el && el.href) {
    //console.log("update home url to:", uri);
    el.href = uri;
  }
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

function doAppPage() {
  // we don't want this in e.g. about page
  document.addEventListener("DOMContentLoaded", start);
}

function doIndexPage() {
  var view = viewGet();
  var loc = window.location.pathname;
  //console.log("doIndexPage(): view:", view, "loc:", loc);
  if (!view) {
    return;
  }
  if (view === "list") {
    if (loc === "/index-grid") {
      window.location = "/";
    }
  } else if (view === "grid") {
    if (loc === "/") {
      window.location = "/index-grid";
    }
  } else {
    console.log("Unknown view:", view);
  }
}

// we don't want to run javascript on about etc. pages
var loc = window.location.pathname;
var isAppPage = loc.indexOf("essential/") != -1;
var isIndexPage = loc === "/" || loc === "/index-grid";

function httpsRedirect() {
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

if (window.g_is_404) {
  do404();
} else if (isIndexPage) {
  doIndexPage();
} else if (isAppPage) {
  doAppPage();
}
updateLinkHome();
httpsRedirect();
