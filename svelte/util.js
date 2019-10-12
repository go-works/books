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
  if (idx === -1) {
    return;
  }
  const tocItem = gBookToc[idx];
  tocItemIdxExpanded.push(idx);
  // console.log(`idx: ${idx}, title: ${tocItem[4]}`)
  idx = item.parentIdx(tocItem);
  setIsExpandedUpwards(idx);
}

export function setTocExpandedForCurrentURL() {
  tocItemIdxExpanded = [];
  const currURI = getLocationLastElementWithHash();
  const n = gBookToc.length;
  let tocItem, uri;
  for (let idx = 0; idx < n; idx++) {
    tocItem = gBookToc[idx];
    uri = tocItemURL(tocItem);
    if (uri === currURI) {
      currentlySelectedIdx.set(idx);
      setIsExpandedUpwards(idx);
      return;
    }
  }
}