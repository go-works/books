<script>
  import Toc from "./Toc.svelte";
  import { item } from "./item.js";
  import { scrollPosSet, currentlySelectedIdx } from "./store.js";
  import { isTocItemExpanded } from "./util.js";

  export let itemIdx = -1;
  export let level = 0;

  const loc = getLocationLastElementWithHash();

  const tocItem = gTocItems[itemIdx];
  const title = item.title(tocItem);
  const url = item.url(tocItem);
  const hasChildren = item.hasChildren(tocItem);

  // let isExpanded = loc.startsWith(url);
  let isExpanded = isTocItemExpanded(itemIdx);

  if (isExpanded) {
    // console.log(`loc: ${loc}, url: ${url}`);
  }

  let isSelected = url === loc;

  currentlySelectedIdx.subscribe(idx => {
    isSelected = idx == itemIdx;
  });

  function toggleExpand() {
    isExpanded = !isExpanded;
    // console.log("toogleExpand(): ", isExpanded);
  }

  function linkClicked() {
    // console.log("link clicked");
    var el = document.getElementById("toc");
    //console.log("el:", el);
    //console.log("scrollTop:", el.scrollTop);
    scrollPosSet(el.scrollTop);
    currentlySelectedIdx.set(itemIdx);
  }
</script>

<div class="toc-item lvl{level}">
  {#if !hasChildren}
    <div class="toc-nav-empty-arrow" />
  {:else if isExpanded}
    <svg class="arrow" on:click={toggleExpand}>
      <use xlink:href="#arrow-expanded" />
    </svg>
  {:else}
    <svg class="arrow" on:click={toggleExpand}>
      <use xlink:href="#arrow-not-expanded" />
    </svg>
  {/if}

  {#if isSelected}
    <b>{title}</b>
  {:else}
    <a class="toc-link" {title} href={url} on:click={linkClicked}>{title}</a>
  {/if}
</div>

{#if hasChildren && isExpanded}
  <Toc parentIdx={itemIdx} level={level + 1} />
{/if}
