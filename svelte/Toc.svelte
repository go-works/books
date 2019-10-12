<script>
  import TocItem from "./TocItem.svelte";
  import { item } from "./item.js";
  import {
    getLocationLastElementWithHash,
    setTocExpandedForCurrentURL
  } from "./util.js";

  export let parentIdx = -1;
  export let level = 0;
  const children = item.childrenForParentIdx(parentIdx);

  if (parentIdx === -1) {
    const loc = getLocationLastElementWithHash();
    // console.log(`loc: ${loc}`);
    setTocExpandedForCurrentURL();
    window.onhashchange = setTocExpandedForCurrentURL;
  }
</script>

{#each children as itemIdx}
  <TocItem {itemIdx} {level} />
{/each}
