package discover

import (
	"context"
	"errors"
	"github.com/chuccp/shareExplorer/core"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Table struct {
	rand      *rand.Rand
	mutex     sync.Mutex
	nodeTable *NodeTable
	localNode *Node
	call      *call
	coreCtx   *core.Context
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (t *Table) ID() ID {
	return t.localNode.id
}
func (t *Table) Ping(node *Node) error {
	return t.call.ping(node.addr)
}
func (t *Table) FindRemoteServer(target ID, node *Node, distances int) (*Node, []*Node, error) {

	return nil, nil, nil
}
func (t *Table) FindServer(target ID, distances int) (*Node, []*Node) {
	return nil, nil
}
func (t *Table) AddNatServer(n *Node) {

}

func (t *Table) addNode(n *Node) {
	t.nodeTable.addNode(n)
}
func (t *Table) addSeedNode(n *Node) {
	t.nodeTable.addSeedNode(n)
}
func (t *Table) queryByFindNode(findNode *FindNode) []*Node {
	return nil
}

func (t *Table) loadAddress() error {
	seeds := make([]*Node, 0)
	addresses, err := t.coreCtx.GetDB().GetAddressModel().QueryAddresses()
	if err != nil {
		return err
	}
	for _, address := range addresses {
		if len(address.Address) > 0 {
			addr, err := net.ResolveUDPAddr("udp", address.Address)
			if err != nil {
				log.Println(err)
				continue
			} else {
				node := &Node{addr: addr, isServer: false, isNatServer: true, isClient: false}
				if len(address.ServerName) >= 64 {
					id, err := StringToId(address.ServerName)
					if err == nil {
						node.id = id
					} else {
						log.Println(err)
					}
				}
				seeds = append(seeds, node)
			}
		}
	}
	for _, seed := range seeds {
		t.addSeedNode(seed)
	}
	if len(seeds) < 1 {
		return errors.New("no node")
	}
	return nil
}

func (t *Table) queryForPage(pageNo, pageSize int) ([]*Node, int) {
	return nil, 0
}

func (t *Table) self() *Node {
	return t.localNode
}

func (t *Table) run() {
	err := t.loadAddress()
	if err != nil {
		log.Panic(err)
		return
	} else {
		t.loop()
	}
}

func (t *Table) loop() {
	var (
		revalidate     = time.NewTimer(t.nextRevalidateTime())
		refresh        = time.NewTimer(t.nextRefreshTime())
		revalidateDone chan struct{}
		refreshDone    = make(chan struct{})
	)
	t.loadAddress()
	go t.doRefresh(refreshDone)
	for {
		select {

		case <-t.ctx.Done():
			{
				break
			}
		case <-refresh.C:
			{
				if refreshDone == nil {
					refreshDone = make(chan struct{})
					go t.doRefresh(refreshDone)
				}
			}
		case <-revalidate.C:
			{
				revalidateDone = make(chan struct{})
				t.doRevalidate(revalidateDone)
			}
		case <-revalidateDone:
			{
				revalidate.Reset(t.nextRevalidateTime())
				revalidateDone = nil
			}
		case <-refreshDone:
			refresh.Reset(t.nextRefreshTime())
		}
	}
}

func (t *Table) nextRevalidateTime() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return time.Duration(t.rand.Int63n(int64(10 * time.Second)))
}

func (t *Table) nextRefreshTime() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	half := 30 * time.Minute / 2
	return half + time.Duration(t.rand.Int63n(int64(half)))
}

func (t *Table) stop() {
	t.ctxCancel()
}

func (t *Table) loadNurseryNodes() {
	if t.nodeTable.hasNurseryNodes() {
		deleteNodes := make([]*Node, 0)
		for _, n := range t.nodeTable.nurseryNodes() {
			if n.ID().IsBlank() {
				node, err := t.call.register(n.addr)
				if err != nil {
					n.errorNum++
					return
				} else {
					t.addNode(node)
					deleteNodes = append(deleteNodes, n)
				}
			}
		}
		t.nodeTable.removeNurseryNodes(deleteNodes)
	}
}

func (t *Table) doRefresh(done chan struct{}) {
	defer close(done)
	t.loadNurseryNodes()
	t.lookup()
}
func (t *Table) lookup() {
	t.lookupSelf()
	for i := 0; i < 3; i++ {
		t.lookupRand()
	}
}
func (t *Table) lookupByTarget(target ID) {
	nodes := t.nodeTable.queryNodesByIdAndDistance(target, findnodeResultLimit)
	for _, n := range nodes.entries {
		nodes, err := t.call.findNode(target, n)
		if err != nil {
			n.errorNum++
			return
		}
		for _, n2 := range nodes {
			t.addNode(n2)
		}
	}
}

func (t *Table) lookupSelf() {
	id := t.localNode.ID()
	t.lookupByTarget(id)
}

func (t *Table) lookupRand() {
	var target ID
	t.rand.Read(target[:])
	t.lookupByTarget(target)
}

func (t *Table) validate(node *Node, index int) {
	value, err := t.call.register(node.addr)
	if err == nil {
		if node.id != value.id {
			t.nodeTable.deleteNode(node)
			t.addNode(value)
			return
		}
		node.liveNessChecks++
		return
	}
	node.errorNum++
	t.nodeTable.replace(index, node)
}

func (t *Table) doRevalidate(done chan struct{}) {
	defer close(done)
	t.loadNurseryNodes()
	node, _, index := t.nodeTable.nodeToRevalidate()
	if node != nil {
		t.validate(node, index)
	}

}
func NewTable2(coreCtx *core.Context, localNode *Node, call *call) *Table {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &Table{ctx: ctx, ctxCancel: ctxCancel, coreCtx: coreCtx, nodeTable: NewNodeTable(localNode), localNode: localNode, call: call}
}
