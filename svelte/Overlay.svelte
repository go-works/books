<script>
  import { createEventDispatcher } from "svelte";

  let overlay;
  const dispatch = createEventDispatcher();

  function dismiss(ev) {
    if (ev.target == overlay) {
      dispatch("wantDismiss");
    }
  }

  function keyDown(ev) {
    console.log("keyDown:", ev);
    /* TODO: also do on ESC */
    console.log("");
  }
</script>

<style>
  .overlay {
    z-index: 10;
    background-color: rgba(0, 0, 0, 0.4);
    position: fixed;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
  }
</style>

<div
  bind:this={overlay}
  class="overlay"
  on:click={dismiss}
  on:keydown={keyDown}>
  <slot />
</div>
