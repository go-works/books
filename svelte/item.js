//export const idxIsExpanded = 0;
const idxURL = 1;
const idxParentIdx = 2;
const idxFirstChildIdx = 3;
const idxTitle = 4;
const idxFirstSynonym = 5;

// TODO: replace references to gBookToc with gTocItems

// index of the parent in the array of all items
function parentIdx(item) {
  return item[idxParentIdx];
}

function hasChildren(item) {
  return item[idxFirstChildIdx] != -1;
}

function parent(item) {
  var idx = parentIdx(item);
  if (idx == -1) {
    return null;
  }
  return gTocItems[idx];
}

// "the-go-command-f2028ab74a354cf2ba6a86acfb813356"
function url(item) {
  while (item) {
    var uri = item[idxURL];
    // toc items that refer to items within the page
    // inherit
    if (uri != "") {
      return uri;
    }
    item = parent(item);
  }
  return "";
}

// all searchable items: title + search synonyms
function searchable(item) {
  return item.slice(idxTitle);
}

function isRoot(item) {
  return parentIdx(item) == -1;
}

function title(item) {
  return item[idxTitle];
}

function firstSynonym(item) {
  return item[idxFirstSynonym];
}

const parentIdxToChildren = {};

const emptyArray = [];

// returns an array of indexes of children in gTocItems
function childrenForParentIdx(parentIdx, firstChildIdx = 0) {
  if (firstChildIdx == -1) {
    // re-use empty array. caller should not modify
    return emptyArray;
  }
  const children = parentIdxToChildren[parentIdx];
  if (children) {
    return children;
  }
  const n = gTocItems.length;
  let res = [];
  for (let i = firstChildIdx; i < n; i++) {
    const tocItem = gTocItems[i];
    if (parentIdx === item.parentIdx(tocItem)) {
      res.push(i);
    }
  }
  parentIdxToChildren[parentIdx] = res;
  return res;
}

function children(item) {
  return childrenForParentIdx(item[idxParentIdx], item[idxFirstChildIdx]);
}

// returns true if has children and some of them articles
// (as opposed to children that are headers within articles)
function hasArticleChildren(item) {
  const idx = item[idxFirstChildIdx];
  if (idx == -1) {
    return false;
  }
  var item = gTocItems[idx];
  var parentIdx = item[idxParentIdx];
  while (idx < gTocItems.length) {
    item = gTocItems[idx];
    if (parentIdx != item[idxParentIdx]) {
      return false;
    }
    var uri = item[idxURL];
    if (uri.indexOf("#") === -1) {
      return true;
    }
    idx += 1;
  }
  return false;
}

export const item = {
  url: url,
  parentIdx: parentIdx,
  parent: parent,
  title: title,
  firstSynonym: firstSynonym,
  children: children,
  childrenForParentIdx: childrenForParentIdx,
  hasChildren: hasChildren,
  searchable: searchable,
  isRoot: isRoot,
  hasArticleChildren: hasArticleChildren,
}