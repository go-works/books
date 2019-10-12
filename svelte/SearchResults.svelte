<script>
  export let results = [];

  /*
  function rebuildSearchResultsUI() {
  var html;
  var results = currentState.searchResults;
  var selectedIdx = currentState.selectedSearchResultIdx;
  var searchWindow = document.getElementById("search-results-window");
  var blurOverlay = document.getElementById("blur-overlay");
  var searchResults = document.getElementById("search-results");
  if (results.length == 0) {
    if (currentSearchTerm == "") {
      searchWindow.style.display = "none";
      blurOverlay.style.display = "none";
    } else {
      searchWindow.style.display = "block";
      blurOverlay.style.display = "block";
      html =
        "<div class='no-search-results'>No search results for '" +
        currentSearchTerm +
        "'</div>";
      searchResults.innerHTML = html;
    }
    return;
  }
  searchWindow.style.display = "block";
  blurOverlay.style.display = "block";
  html = buildResultsHTML(results, selectedIdx);
  searchResults.innerHTML = html;

  // ensure element is scrolled into view
  window.requestAnimationFrame(() => {
    if (selectedIdx < 0) {
      return;
    }
    var id = "search-result-no-" + selectedIdx;
    var el = document.getElementById(id);
    scrollElementIntoView(el, true);
  });
}

var searchInOpt = {
  cls: "in"
};
*/

  /* results is array of items:
{
  tocItem: [],
  term: "",
  match: [[idx, len], ...],
}
*/
  /*
function buildResultsHTML(results, selectedIdx) {
  var a = [];
  var n = results.length;
  for (var i = 0; i < n; i++) {
    var r = results[i];
    var tocItem = r.tocItem;
    var term = r.term;
    var matches = r.match;

    var html = hilightSearchResult(term, matches);
    // TODO: get multi-level path (e.g. for 'json' where in Refelection / Uses for reflection chapter)
    var inTxt = getArticlePath(tocItem, term);
    if (!inTxt) {
      inTxt = getParentTitle(tocItem);
    }
    if (inTxt) {
      html += " " + inTagRaw("span", inTxt, searchInOpt);
    }

    var opt = {
      id: "search-result-no-" + i,
      classes: ["search-result"]
    };
    if (i == selectedIdx) {
      opt.classes.push("search-result-selected");
    }
    var s = div(html, opt);
    a.push(s);
  }
  return a.join("\n");
}
*/
</script>
