package discover

import (
	"encoding/json"
	"github.com/chuccp/shareExplorer/core"
	"math/rand"
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
func (ts *TableStore) AddTable(localNode *LocalNode) {
	ts.tableStore.LoadOrStore(localNode.id, NewTable(ts.context, localNode))
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
	rand       *rand.Rand
	tableStore *TableStore
}

func NewTableGroup(context *core.Context) *TableGroup {
	return &TableGroup{config: context.GetServerConfig().GetConfig(), context: context, tableStore: NewTableStore(context)}
}

func (tableGroup *TableGroup) loop() {
	var (
		revalidate     = time.NewTimer(tableGroup.nextRevalidateTime())
		revalidateDone chan struct{}
	)

	for {
		select {
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

		}
	}
}
func (tableGroup *TableGroup) nextRevalidateTime() time.Duration {
	tableGroup.mutex.Lock()
	defer tableGroup.mutex.Unlock()
	return time.Duration(tableGroup.rand.Int63n(int64(10 * time.Second)))
}

func (tableGroup *TableGroup) doRevalidate(done chan<- struct{}) {

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

func (tableGroup *TableGroup) AddTable(localNode *LocalNode) {
	tableGroup.tableStore.AddTable(localNode)
}

type Table struct {
	buckets    [nBuckets]*bucket
	nursery    []*node //bootstrap nodes
	context    *core.Context
	httpClient *core.HttpClient
	localNode  *LocalNode
}

func (tab *Table) ping(n *Node) {
	data, _ := json.Marshal(tab.localNode)
	_, err := tab.httpClient.PostRequest(n.addr.String(), "/discover/ping", string(data))
	if err != nil {
		return
	}
}
func (tab *Table) addSeenNode(n *node) {

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
func (tab *Table) findNode(n *Node, distances []uint) {

}

func NewTable(context *core.Context, localNode *LocalNode) *Table {

	table := &Table{
		context:    context,
		httpClient: core.NewHttpClient(context),
		localNode:  localNode,
	}
	for i := range table.buckets {
		table.buckets[i] = &bucket{
			index: i,
			ips:   DistinctNetSet{Subnet: bucketSubnet, Limit: bucketIPLimit},
		}
	}

	return table
}
