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

function setState(newState, now = false) {
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
}

function onClick(ev) {
  var el = ev.target;
  //console.log("onClick ev:", ev, "el:", el);

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
}

function onEnter(ev) {
}

function onEscape(ev) {
}

function onUpDown(ev) {
}

function onKeyDown(ev) {
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
