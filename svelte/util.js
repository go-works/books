import { item } from "./item.js";
import { currentlySelectedIdx } from "./store.js";

export function getLocationLastElement() {
  var loc = window.location.pathname;
  var parts = loc.split("/");
  var lastIdx = parts.length - 1;
  return parts[lastIdx];
}

export function getLocationLastElementWithHash() {
  var loc = window.location.pathname;
  var parts = loc.split("/");
  var lastIdx = parts.length - 1;
  return parts[lastIdx] + window.location.hash;
}

// TODO: maybe move to item.js
// remembers which toc items are expanded, by their index
let tocItemIdxExpanded = [];

export function isTocItemExpanded(idx) {
  for (let i of tocItemIdxExpanded) {
    if (i === idx) {
      return true;
    }
  }
  return false;
}

function setIsExpandedUpwards(idx) {
  const tocItem = gTocItems[idx];
  tocItemIdxExpanded.push(idx);
  // console.log(`idx: ${idx}, title: ${tocItem[4]}`)
  const newIdx = item.parentIdx(tocItem);
  if (newIdx != -1) {
    setIsExpandedUpwards(newIdx);
  }
}

export function setTocExpandedForCurrentURL() {
  tocItemIdxExpanded = [];
  const currURI = getLocationLastElementWithHash();
  const n = gTocItems.length;
  let tocItem, uri;
  for (let idx = 0; idx < n; idx++) {
    tocItem = gTocItems[idx];
    uri = item.url(tocItem);
    if (uri === currURI) {
      currentlySelectedIdx.set(idx);
      setIsExpandedUpwards(idx);
      return idx;
    }
  }
  return 0;
}

// return true if this is Esc key event
export function isEsc(ev) {
  // TODO: optimize, check code
  // Esc is Edge
  return (ev.key == "Escape") || (ev.key == "Esc");
}

export function isEnter(ev) {
  return ev.key == "Enter";
}

export function isUp(ev) {
  return (ev.key == "ArrowUp") || (ev.key == "Up");
}

export function isDown(ev) {
  return (ev.key == "ArrowDown") || (ev.key == "Down");
}

// navigation up is: Up or Ctrl-P
export function isNavUp(ev) {
  if (isUp(ev)) {
    return true;
  }
  return ev.ctrlKey && (ev.keyCode === 80);
}

// navigation down is: Down or Ctrl-N
export function isNavDown(ev) {
  if (isDown(ev)) {
    return true;
  }
  return ev.ctrlKey && (ev.keyCode === 78);
}

// returns a debouncer function. Usage:
// var debouncer = makeDebouncer(250);
// function fn() { ... }
// debouncer(fn)
export function makeDebouncer(timeInMs) {
  let interval;
  return function (f) {
    clearTimeout(interval);
    interval = setTimeout(() => {
      interval = null;
      f();
    }, timeInMs);
  };
}

// https://github.com/Treora/scroll-into-view/blob/master/polyfill.js
// TODO: passing options = { center: true } doesn't work
// TODO: jusst use el.scrollIntoView?
export function scrollElementIntoView(el, options) {
  // Use traditional scrollIntoView when traditional argument is given.
  if (options === undefined || options === true || options === false) {
    el.scrollIntoView(options);
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

