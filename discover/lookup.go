package discover

import (
	"sort"
)

type nodesByDistance struct {
	entries []*Node
	target  ID
}

// push adds the given node to the list, keeping the total size below maxElems.
func (h *nodesByDistance) push(n *Node, maxElems int) {
	ix := sort.Search(len(h.entries), func(i int) bool {
		return DistCmp(h.target, h.entries[i].ID(), n.ID()) > 0
	})

	end := len(h.entries)
	if len(h.entries) < maxElems {
		h.entries = append(h.entries, n)
	}
	if ix < end {
		// Slide existing entries down to make room.
		// This will overwrite the entry we just appended.
		copy(h.entries[ix+1:], h.entries[ix:])
		h.entries[ix] = n
	}
}
