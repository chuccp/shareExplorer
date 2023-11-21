package discover

import (
	"encoding/json"
	"github.com/chuccp/shareExplorer/core"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	nBuckets                    = 17
	bucketMinDistance           = 239
	bucketIPLimit, bucketSubnet = 2, 24 //
)

type node struct {
	Node
	addedAt        time.Time
	liveNessChecks uint
}

func wrapNode(n *Node) *node {
	return &node{Node: *n}
}

type bucket struct {
	entries      []*node
	replacements []*node
	index        int
	ips          DistinctNetSet
}
type TableStore struct {
	tableStore *sync.Map
	context    *core.Context
}

func NewTableStore(context *core.Context) *TableStore {
	return &TableStore{tableStore: new(sync.Map), context: context}
}
func (ts *TableStore) AddTable(localNode *LocalNode) *Table {
	table := NewTable(ts.context, localNode)
	t, _ := ts.tableStore.LoadOrStore(localNode.id, table)
	return t.(*Table)
}
func (ts *TableStore) RangeTable(f func(key ID, value *Table) bool) {
	ts.tableStore.Range(func(key, value any) bool {
		return f(key.(ID), value.(*Table))
	})
}

type TableGroup struct {
	mutex      sync.Mutex
	config     *core.Config
	context    *core.Context
	tableStore *TableStore
	rand       *rand.Rand
}

func NewTableGroup(context *core.Context) *TableGroup {
	return &TableGroup{config: context.GetServerConfig().GetConfig(), rand: rand.New(rand.NewSource(0)), context: context, tableStore: NewTableStore(context)}
}
func (tableGroup *TableGroup) doRefresh(done chan struct{}) {
	defer close(done)
	tableGroup.tableStore.RangeTable(func(key ID, value *Table) bool {
		value.doRefresh()
		return true
	})
}
func (tableGroup *TableGroup) loop() {
	var (
		revalidate     = time.NewTimer(tableGroup.nextRevalidateTime())
		refresh        = time.NewTimer(tableGroup.nextRefreshTime())
		revalidateDone chan struct{}
		refreshDone    = make(chan struct{})
	)
	go tableGroup.doRefresh(refreshDone)
	for {
		select {
		case <-refresh.C:
			{
				if refreshDone == nil {
					refreshDone = make(chan struct{})
					go tableGroup.doRefresh(refreshDone)
				}
			}
		case <-revalidate.C:
			{
				revalidateDone = make(chan struct{})
				tableGroup.doRevalidate(revalidateDone)
			}
		case <-revalidateDone:
			{
				revalidate.Reset(tableGroup.nextRevalidateTime())
				revalidateDone = nil
			}
		case <-refreshDone:

			refresh.Reset(tableGroup.nextRefreshTime())

		}
	}
}

func (tableGroup *TableGroup) nextRevalidateTime() time.Duration {
	tableGroup.mutex.Lock()
	defer tableGroup.mutex.Unlock()
	return time.Duration(tableGroup.rand.Int63n(int64(10 * time.Second)))
}
func (tableGroup *TableGroup) nextRefreshTime() time.Duration {
	tableGroup.mutex.Lock()
	defer tableGroup.mutex.Unlock()
	half := 30 * time.Minute / 2
	return half + time.Duration(tableGroup.rand.Int63n(int64(half)))
}

func (tableGroup *TableGroup) doRevalidate(done chan<- struct{}) {
	defer close(done)
	tableGroup.tableStore.RangeTable(func(key ID, value *Table) bool {
		value.register()
		return true
	})
}

func (tableGroup *TableGroup) run() {
	go tableGroup.loop()
}
func (tableGroup *TableGroup) addSeenNode(n *node) {
	tableGroup.mutex.Lock()
	defer tableGroup.mutex.Unlock()
	tableGroup.tableStore.RangeTable(func(key ID, value *Table) bool {
		if n.id != key {
			value.addSeenNode(n)
		}
		return true
	})
}

func (tableGroup *TableGroup) AddTable(localNode *LocalNode) *Table {
	return tableGroup.tableStore.AddTable(localNode)
}

type Table struct {
	buckets    [nBuckets]*bucket
	nursery    []*node //bootstrap nodes
	context    *core.Context
	httpClient *core.HttpClient
	localNode  *LocalNode
	rand       *rand.Rand
}

func (tab *Table) addNursery(addr *net.UDPAddr) {
	if !containsAddress(tab.nursery, addr) {
		n := wrapNode(NewNursery(addr))
		n.addedAt = time.Now()
		tab.nursery = append(tab.nursery, n)
	}
}

func containsAddress(ns []*node, addr *net.UDPAddr) bool {
	for _, v := range ns {
		if IsSameAddress(v.addr, addr) {
			return true
		}
	}
	return false
}

func (tab *Table) addSeenNode(n *node) {

}
func (tab *Table) doRefresh() {
	tab.loadSeedNodes()
	tab.lookupSelf()
}
func (tab *Table) lookupSelf() {

}

func (tab *Table) newLookup(ctx *core.Context, target ID) {
	newLookup(tab, target, ctx).run()
}

func (tab *Table) loadSeedNodes() {

}

func contains(ns []*node, id ID) bool {
	for _, n := range ns {
		if n.ID() == id {
			return true
		}
	}
	return false
}

func (tab *Table) bucket(id ID) *bucket {
	d := LogDist(tab.localNode.id, id)
	return tab.bucketAtDistance(d)
}
func (tab *Table) bucketAtDistance(d int) *bucket {
	if d <= bucketMinDistance {
		return tab.buckets[0]
	}
	return tab.buckets[d-bucketMinDistance-1]
}
func (tab *Table) nodeToRevalidate() (n *node, bi int) {
	for _, bi = range tab.rand.Perm(len(tab.buckets)) {
		b := tab.buckets[bi]
		if len(b.entries) > 0 {
			last := b.entries[len(b.entries)-1]
			return last, bi
		}
	}
	return nil, 0
}

func (tab *Table) registerNursery() {
	for _, n := range tab.nursery {

		n.ID()
	}

}

func (tab *Table) register() {

	node, _ := tab.nodeToRevalidate()
	tab.register0(node)
}

func (tab *Table) register0(node *node) {
	data, _ := json.Marshal(tab.localNode)
	_, err := tab.httpClient.PostRequest(node.addr.String(), "/discover/register", string(data))
	if err != nil {
		return
	}
}

func (tab *Table) findNode(n *Node, distances []uint) {

}

func NewTable(context *core.Context, localNode *LocalNode) *Table {

	table := &Table{
		context:    context,
		httpClient: core.NewHttpClient(context),
		localNode:  localNode,
		rand:       rand.New(rand.NewSource(0)),
	}
	for i := range table.buckets {
		table.buckets[i] = &bucket{
			index: i,
			ips:   DistinctNetSet{Subnet: bucketSubnet, Limit: bucketIPLimit},
		}
	}

	return table
}
