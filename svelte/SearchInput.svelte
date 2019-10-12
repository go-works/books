<script>
  import { onMount, onDestroy } from "svelte";
  import { search } from "./search.js";

  export let bookTitle = "";
  let inputSearchTerm = "";
  let input;

  console.log("SearchInput");

  $: doSearch(inputSearchTerm);

  function onKeyDown(ev) {
    if (ev.key == "/") {
      input.focus();
      ev.preventDefault();
      return;
    }

    // Esc is Edge
    if (ev.key == "Escape" || ev.key == "Esc") {
      inputSearchTerm = "";
      input.blur();
      return;
    }
  }

  onMount(() => {
    document.addEventListener("keydown", onKeyDown);
  });

  onDestroy(() => {
    document.removeEventListener("keydown", onKeyDown);
  });

  function doSearch(searchTerm) {
    // TODO: debounce search
    searchTerm = searchTerm.trim().toLowerCase();
    if (searchTerm.length == 0) {
      return;
    }
    console.log(`doSearch: '${searchTerm}'`);
    const res = search(searchTerm);
    console.log("res:", res);
  }

  /*
function doSearch(searchTerm) {

// console.log("search results:", res);
  setState({
    searchResults: res,
    selectedSearchResultIdx: 0
  });
}
*/
</script>

<style>
  input {
    width: 100%;
    font-size: 16px;
    padding: 2px 8px;
    background-color: white;
    /* filter: opacity(1); */
    /* border-color: #717274; */
    /* box-shadow: inset 0 1px 1px rgba(0,0,0,.075); */
    border: 1px solid silver;
    /* box-shadow: inset 0 0 0 0 transparent; */
    outline: 0;
    z-index: 25;
  }

  input:hover {
    border-color: #a0a0a0;
  }

  input::placeholder {
    color: #aaaaaa;
  }

  /* trick to make placeholder invisible when input field is focused */
  input:focus::placeholder {
    color: white;
  }

  /* no blue border when focused */
  input:focus {
    /* border: 1px solid lightskyblue; */
    /* border: 1px solid #aaaaaa; */
    border-color: #a0a0a0;
    box-shadow: inset 0 1px 1px rgba(0, 0, 0, 0.075);
  }
</style>

<input
  placeholder="Search '{bookTitle}' Tip: press '/'."
  bind:value={inputSearchTerm}
  bind:this={input} />
