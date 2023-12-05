package discover

import (
	"github.com/chuccp/shareExplorer/core"
	"sort"
)

type queryFunc func(*node) ([]*node, error)

type lookup struct {
	tab       *Table
	context   *core.Context
	result    nodesByDistance
	target    ID
	queryFunc queryFunc
}

func newLookup(tab *Table, target ID, context *core.Context, queryFunc queryFunc) *lookup {
	return &lookup{tab: tab, target: target, context: context, queryFunc: queryFunc}
}
func (l *lookup) run() *lookup {

	nodesByDistance := l.tab.findNodeByID(l.target, bucketSize, false)
	for _, entry := range nodesByDistance.entries {
		go l.queryFunc(entry)
	}
	return l
}

type nodesByDistance struct {
	entries []*node
	target  ID
}

// push adds the given node to the list, keeping the total size below maxElems.
func (h *nodesByDistance) push(n *node, maxElems int) {
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
