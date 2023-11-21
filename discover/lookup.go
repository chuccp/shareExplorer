package discover

import "github.com/chuccp/shareExplorer/core"

type lookup struct {
	tab       *Table
	context   *core.Context
	queryfunc func(*node) ([]*node, error)
}

func newLookup(tab *Table, target ID, context *core.Context) *lookup {
	return &lookup{tab: tab, context: context}
}
func (l *lookup) run() {

}
