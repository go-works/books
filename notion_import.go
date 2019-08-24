package main

import (
	"fmt"

	"github.com/kjk/notionapi"
)

// convert 2131b10c-ebf6-4938-a127-7089ff02dbe4 to 2131b10cebf64938a1277089ff02dbe4
// TODO: replace with direct use of notionapi.ToNoDashID
func toNoDashID(id string) string {
	return notionapi.ToNoDashID(id)
}

func updateFormatIfNeeded(page *notionapi.Page) bool {
	// can't write back without a token
	if notionAuthToken == "" {
		return false
	}
	args := map[string]interface{}{}
	format := page.Root().FormatPage()
	if format == nil || !format.PageSmallText {
		args["page_small_text"] = true
	}
	if format == nil || !format.PageFullWidth {
		args["page_full_width"] = true
	}
	if len(args) == 0 {
		return false
	}
	fmt.Printf("  updating format to %v\n", args)
	err := page.SetFormat(args)
	if err != nil {
		fmt.Printf("updateFormatIfNeeded: page.SetFormat() failed with '%s'\n", err)
	}
	return true
}

func updateTitleIfNeeded(page *notionapi.Page) bool {
	// can't write back without a token
	if notionAuthToken == "" {
		return false
	}
	newTitle := cleanTitle(page.Root().Title)
	if newTitle == page.Root().Title {
		return false
	}
	fmt.Printf("  updating title to '%s'\n", newTitle)
	err := page.SetTitle(newTitle)
	if err != nil {
		fmt.Printf("updateTitleIfNeeded: page.SetTitle() failed with '%s'\n", err)
	}
	return true
}

func updateFormatOrTitleIfNeeded(page *notionapi.Page) bool {
	updated1 := updateFormatIfNeeded(page)
	updated2 := updateTitleIfNeeded(page)
	return updated1 || updated2
}

func isIDEqual(id1, id2 string) bool {
	return notionapi.ToNoDashID(id1) == notionapi.ToNoDashID(id2)
}
