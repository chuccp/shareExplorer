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

type Table2 struct {
	rand      *rand.Rand
	mutex     sync.Mutex
	nodeTable *NodeTable
	localNode *Node
	call      *call
	coreCtx   *core.Context
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (t *Table2) ID() ID {
	return t.localNode.id
}
func (t *Table2) Ping(node *Node) error {
	return t.call.ping(node.addr)
}
func (t *Table2) FindRemoteServer(target ID, node *Node, distances int) (*Node, []*Node, error) {

	return nil, nil, nil
}
func (t *Table2) FindServer(target ID, distances int) (*Node, []*Node) {
	return nil, nil
}
func (t *Table2) AddNatServer(n *node) {

}

func (t *Table2) addNode(n *Node) {
	t.nodeTable.addNode(n)
}
func (t *Table2) addSeedNode(n *Node) {
	t.nodeTable.addSeedNode(n)
}
func (t *Table2) queryByFindNode(findNode *FindNode) []*Node {
	return nil
}

func (t *Table2) loadAddress() error {
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

func (t *Table2) queryForPage(pageNo, pageSize int) ([]*Node, int) {
	return nil, 0
}

func (t *Table2) self() *Node {
	return t.localNode
}

func (t *Table2) run() {
	err := t.loadAddress()
	if err != nil {
		log.Panic(err)
		return
	} else {
		t.loop()
	}
}

func (t *Table2) loop() {
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

func (t *Table2) nextRevalidateTime() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return time.Duration(t.rand.Int63n(int64(10 * time.Second)))
}

func (t *Table2) nextRefreshTime() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	half := 30 * time.Minute / 2
	return half + time.Duration(t.rand.Int63n(int64(half)))
}

func (t *Table2) stop() {
	t.ctxCancel()
}

func (t *Table2) loadNurseryNodes() {
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

func (t *Table2) doRefresh(done chan struct{}) {
	defer close(done)
	t.loadNurseryNodes()
	t.lookup()
}
func (t *Table2) lookup() {
	t.lookupSelf()
	for i := 0; i < 3; i++ {
		t.lookupRand()
	}
}
func (t *Table2) lookupByTarget(target ID) {
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

func (t *Table2) lookupSelf() {
	id := t.localNode.ID()
	t.lookupByTarget(id)
}

func (t *Table2) lookupRand() {
	var target ID
	t.rand.Read(target[:])
	t.lookupByTarget(target)
}

func (t *Table2) validate(node *Node, index int) {
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

func (t *Table2) doRevalidate(done chan struct{}) {
	defer close(done)
	t.loadNurseryNodes()
	node, _, index := t.nodeTable.nodeToRevalidate()
	if node != nil {
		t.validate(node, index)
	}

}
func NewTable2(coreCtx *core.Context, localNode *Node, call *call) *Table2 {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &Table2{ctx: ctx, ctxCancel: ctxCancel, coreCtx: coreCtx, nodeTable: NewNodeTable(localNode), localNode: localNode, call: call}
}
