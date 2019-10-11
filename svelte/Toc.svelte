<script>
  import TocItem from "./TocItem.svelte";

  export let toc = [];
</script>

{#each toc as tocItem}
  <TocItem item={tocItem} level={0} />
{/each}
