package discover

import (
	"encoding/json"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/web"
	"math/rand"
	"sync"
	"time"
)

const (
	nBuckets = 17
)

type node struct {
	addedAt        time.Time
	liveNessChecks uint
}

type bucket struct {
	entries      []*node
	replacements []*node
	index        int
}
type TableStore struct {
	tableStore *sync.Map
	context    *core.Context
}

func NewTableStore(context *core.Context) *TableStore {
	return &TableStore{tableStore: new(sync.Map)}
}
func (ts *TableStore) AddTable(localNode *LocalNode) {
	ts.tableStore.LoadOrStore(localNode.id, NewTable(ts.context, localNode))
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

func (tableGroup *TableGroup) AddTable(localNode *LocalNode) {
	tableGroup.tableStore.AddTable(localNode)
}

type Table struct {
	buckets    [nBuckets]*bucket
	nursery    []*node //bootstrap nodes
	context    *core.Context
	httpClient *web.HttpClient
	localNode  *LocalNode
}

func (tab *Table) ping(n *Node) {
	data, _ := json.Marshal(tab.localNode)
	_, err := tab.httpClient.PostRequest(n.addr.String(), "/discover/ping", string(data))
	if err != nil {
		return
	}
}

func (tab *Table) findNode(n *Node, distances []uint) {

}

func NewTable(context *core.Context, localNode *LocalNode) *Table {
	return &Table{
		context:    context,
		httpClient: web.NewHttpClient(context),
		localNode:  localNode,
	}
}
