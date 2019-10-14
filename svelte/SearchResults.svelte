<script>
  import Overlay from "./Overlay.svelte";
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import { isEnter, isNavUp, isNavDown } from "./util.js";
  import { item } from "./item.js";

  const dispatch = createEventDispatcher();

  /* results is array of items:
    {
      tocItem: [],
      term: "",
      match: [[idx, len], ...],
    }
  */
  export let results = [];
  export let selectedIdx = 0;
  export let searchTerm = "";

  let selectedElement;

  $: selectedElementChanged(selectedElement);

  function selectedElementChanged(el) {
    if (!el) {
      return;
    }
    // TODO: test on Safari
    // https://developer.mozilla.org/en-US/docs/Web/API/Element/scrollIntoView
    el.scrollIntoView(false);
  }

  // must add them globally to be called even when search
  // input field has focus
  onMount(() => {
    document.addEventListener("keydown", keyDown);
  });

  onDestroy(() => {
    document.removeEventListener("keydown", keyDown);
  });

  function getParentTitle(tocItem) {
    var res = "";
    var parent = item.parent(tocItem);
    while (parent) {
      var s = item.title(parent);
      if (res) {
        s = s + " / ";
      }
      res = s + res;
      parent = item.parent(parent);
    }
    return res;
  }

  // return true if term is a search synonym inside tocItem
  function isMatchSynonym(tocItem, term) {
    term = term.toLowerCase();
    var title = item.title(tocItem).toLowerCase();
    return title != term;
  }

  // if search matched synonym returns "${chapterTitle} / ${articleTitle}"
  // otherwise empty string
  function getArticlePath(tocItem, term) {
    if (!isMatchSynonym(tocItem, term)) {
      return null;
    }
    var title = item.title(tocItem);
    var parentTitle = getParentTitle(tocItem);
    if (parentTitle == "") {
      return title;
    }
    return parentTitle + " / " + title;
  }

  function getWhere(idx) {
    const r = results[idx];
    var tocItem = r.tocItem;
    var term = r.term;

    // TODO: get multi-level path (e.g. for 'json' where in Refelection / Uses for reflection chapter)
    const inTxt = getArticlePath(tocItem, term);
    if (inTxt) {
      return inTxt;
    }
    return getParentTitle(tocItem);
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
        res += `<span>${s}</span>`;
      }
      s = txt.substring(idx, idx + len);
      res += `<span class="hili">${s}</span>`;
      prevIdx = idx + len;
    }
    var txtLen = txt.length;
    nonHilightLen = txtLen - prevIdx;
    if (nonHilightLen > 0) {
      s = txt.substring(prevIdx, prevIdx + nonHilightLen);
      res += `<span>${s}</span>`;
    }
    return res;
  }

  function hiliHTML(idx) {
    const r = results[idx];
    // console.log("hili: idx:", idx, "r:", r);
    return hilightSearchResult(r.term, r.match);
  }

  function isChapterOrArticleURL(s) {
    var isChapterOrArticle = s.indexOf("#") === -1;
    return isChapterOrArticle;
  }

  function navigateToSearchResult(idx) {
    // console.log("navigateToSearchResult:", idx);
    var loc = window.location.pathname;
    var parts = loc.split("/");
    var lastIdx = parts.length - 1;
    var lastURL = parts[lastIdx];
    var selected = results[idx];
    var tocItem = selected.tocItem;

    // either replace chapter/article url or append to book url
    var uri = item.url(tocItem);
    if (isChapterOrArticleURL(lastURL)) {
      parts[lastIdx] = uri;
    } else {
      parts.push(uri);
    }
    loc = parts.join("/");
    window.location = loc;
    dispatch("wantDismiss");
  }

  function dir(ev) {
    if (isNavUp(ev)) {
      return -1;
    }
    if (isNavDown(ev)) {
      return 1;
    }
    return 0;
  }

  function keyDown(ev) {
    // console.log("SearchResults.keyDown:", ev);
    if (isEnter(ev)) {
      navigateToSearchResult(selectedIdx);
      ev.stopPropagation();
      return;
    }
    const n = dir(ev);
    if (n === 0) {
      return;
    }
    selectedIdx += n;
    if (selectedIdx < 0) {
      selectedIdx = 0;
    }
    const lastIdx = results.length - 1;
    if (selectedIdx > lastIdx) {
      selectedIdx = lastIdx;
    }
    // console.log("newSelected", selectedIdx);
    ev.stopPropagation();
  }

  function clicked(idx) {
    // console.log("clicked:", idx);
    navigateToSearchResult(idx);
  }
</script>

<style>
  .search-results-window {
    position: fixed;
    top: 28px;
    width: 74vw;
    left: 13vw; /* (100 - 74) / 2 */
    right: 13vw;

    border: 1px solid #aaaaaa;
    z-index: 25;

    /* min-height: 320px; */
    background-color: white;
  }

  .search-results {
    max-height: 70vh;
    padding: 4px 8px;
    line-height: 1.3em;
    cursor: pointer;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .search-result:hover {
    background-color: #eeeeee;
  }

  .search-results-help {
    color: #717274;
    background-color: #f9f9f9;
    padding: 8px 8px;
    font-size: 0.7em;
  }

  .search-result-selected {
    background-color: #eeeeee;
  }

  .no-search-results {
    padding-top: 48px;
    padding-bottom: 48px;
    margin-left: auto;
    margin-right: auto;
    /* background-color: #dddddd; */
    text-align: center;
  }

  @media screen and (max-width: 500px) {
    .search-results-window {
      width: 90vw;
      left: 5vw; /* (100 - 90) / 2 */
      right: 5vw;
    }

    /* leave space for a on-screen keyboard. Tested on
     Android Pixel device */
    .search-results {
      max-height: 70vh;
    }
  }

  /* higlight search results with yellow-ish background */
  :global(.hili) {
    /* font-weight: bold; */
    /*padding: 1px 2px; */
    /* background: #ffeb3b; */
    background: rgba(255, 235, 59, 0.6);
    /* border-radius: 2px; */
    /* font-weight: bold; */
    /* background-color: lightskyblue; */
  }
</style>

<Overlay>
  <div class="search-results-window">
    <div class="search-results">
      {#if results.length === 0}
        <div class="no-search-results">No search results for {searchTerm}</div>
      {/if}
      {#each results as r, idx (r.term)}
        {#if idx === selectedIdx}
          <div
            bind:this={selectedElement}
            on:click={() => clicked(idx)}
            class="search-result search-result-selected">
            {@html hiliHTML(idx)}
            <span class="in">{getWhere(idx)}</span>
          </div>
        {:else}
          <div class="search-result" on:click={() => clicked(idx)}>
            {@html hiliHTML(idx)}
            <span class="in">{getWhere(idx)}</span>
          </div>
        {/if}
      {/each}
    </div>
    <div class="search-results-help">
      &nbsp;&nbsp;&uarr; &darr; to navigate &nbsp;&nbsp;&nbsp; &crarr; to select
      &nbsp;&nbsp;&nbsp; Esc to close
    </div>
  </div>
</Overlay>
