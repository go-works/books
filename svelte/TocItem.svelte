<script>
  export let item = [];
  export let level = 0;

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
  function title() {
    return tocItemTitle(item);
  }
  function hasChildren() {
    return tocItemHasChildren(item);
  }
  function isExpanded() {
    return false;
  }
</script>

<div class="toc-item lvl{level}" id="ti-10">
  {#if hasChildren()}
    <div class="toc-nav-empty-arrow" />
  {:else if isExpanded()}
    <svg class="arrow">
      <use xlink:href="#arrow-expanded" />
    </svg>
  {:else}
    <svg class="arrow">
      <use xlink:href="#arrow-not-expanded" />
    </svg>
  {/if}
  {title()}
</div>
