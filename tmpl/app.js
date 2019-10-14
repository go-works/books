
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

if (isIndexPage) {
  doIndexPage();
  updateLinkHome();
}
