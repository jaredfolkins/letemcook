package models

import "sort"

// ReorderPagesSequentially sorts pages by PageID and reassigns them sequential IDs starting from 1
func ReorderPagesSequentially(pages []Page) []Page {
	if len(pages) == 0 {
		return pages
	}

	// Sort pages by their current PageID
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].PageID < pages[j].PageID
	})

	// Reassign PageIDs sequentially starting from 1
	for i := range pages {
		pages[i].PageID = i + 1
	}

	return pages
}
