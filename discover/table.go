package discover

import (
	"context"
	"errors"
	"github.com/chuccp/shareExplorer/core"
	"go.uber.org/zap"
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

func (table *Table) ID() ID {
	return table.localNode.id
}
func (table *Table) Ping(node *Node) error {
	return table.call.ping(node.addr)
}
func (table *Table) FindRemoteServer(target ID, node *Node, distances int) (*Node, []*Node, error) {
	return table.call.findServer(target, distances, node.addr)
}
func (table *Table) FindServer(target ID, distances int) (*Node, []*Node) {
	if table.localNode.id == target {
		return table.localNode, []*Node{}
	}
	node, fa := table.nodeTable.queryServer(target)
	table.coreCtx.GetLog().Debug("queryServer", zap.Any("node", node), zap.Bool("fa", fa))
	return node, table.nodeTable.collectTableFindNode(distances)
}
func (table *Table) AddNatServer(n *Node) {
	table.nodeTable.addNode(n)
}
func (table *Table) addNode(n *Node) {
	table.nodeTable.addNode(n)
}
func (table *Table) addSeedNode(n *Node) {
	table.nodeTable.addSeedNode(n)
}
func (table *Table) FindNode(findNode *FindNode) []*Node {
	return table.nodeTable.collectTableNodes(findNode.addr.IP, findNode.Distances, findnodeResultLimit)
}

func (table *Table) loadAddress() error {
	seeds := make([]*Node, 0)
	addresses, err := table.coreCtx.GetDB().GetAddressModel().QueryAddresses()
	if err != nil {
		table.coreCtx.GetLog().Error("loadAddress", zap.Error(err))
		return err
	}
	for _, address := range addresses {
		if len(address.Address) > 0 {
			addr, err := net.ResolveUDPAddr("udp", address.Address)
			if err != nil {
				table.coreCtx.GetLog().Error("loadAddress", zap.Error(err))
				continue
			} else {
				node := &Node{addr: addr, isServer: false, isNatServer: true, isClient: false}
				if len(address.ServerName) >= 64 {
					id, err := StringToId(address.ServerName)
					if err == nil {
						node.id = id
					} else {
						table.coreCtx.GetLog().Error("loadAddress", zap.Error(err))
					}
				}
				seeds = append(seeds, node)
			}
		}
	}
	for _, seed := range seeds {
		table.addSeedNode(seed)
	}
	if len(seeds) < 1 {
		err = errors.New("db no node")
		table.coreCtx.GetLog().Error("loadAddress", zap.Error(err))
		return err
	}
	return nil
}

func (table *Table) queryNatServerForPage(pageNo, pageSize int) ([]*Node, int) {

	return table.nodeTable.queryNatServerForPage(pageNo, pageSize)
}

func (table *Table) self() *Node {
	return table.localNode
}

func (table *Table) run() {
	table.mutex.Lock()
	if table.ctxCancel != nil {
		table.ctxCancel()
		table.ctxCancel = nil
	}
	table.ctx, table.ctxCancel = context.WithCancel(context.Background())
	defer table.mutex.Unlock()
	err := table.loadAddress()
	if err != nil {
		return
	} else {
		go table.loop(table.ctx)
	}
}

func (table *Table) loop(ctx context.Context) {
	var (
		revalidate     = time.NewTimer(table.nextRevalidateTime())
		refresh        = time.NewTimer(table.nextRefreshTime())
		revalidateDone chan struct{}
		refreshDone    = make(chan struct{})
	)
	go table.doRefresh(refreshDone)
	for {
		select {

		case <-ctx.Done():
			{
				break
			}
		case <-refresh.C:
			{
				if refreshDone == nil {
					refreshDone = make(chan struct{})
					go table.doRefresh(refreshDone)
				}
			}
		case <-revalidate.C:
			{
				revalidateDone = make(chan struct{})
				table.doRevalidate(revalidateDone)
			}
		case <-revalidateDone:
			{
				revalidate.Reset(table.nextRevalidateTime())
				revalidateDone = nil
			}
		case <-refreshDone:
			refresh.Reset(table.nextRefreshTime())
		}
	}
}

func (table *Table) nextRevalidateTime() time.Duration {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	return time.Duration(table.rand.Int63n(int64(10 * time.Second)))
}

func (table *Table) nextRefreshTime() time.Duration {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	half := 30 * time.Minute / 2
	return half + time.Duration(table.rand.Int63n(int64(half)))
}

func (table *Table) stop() {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	if table.ctxCancel != nil {
		table.ctxCancel()
		table.ctxCancel = nil
	}
}

func (table *Table) loadNurseryNodes() {
	if table.nodeTable.hasNurseryNodes() {
		deleteNodes := make([]*Node, 0)
		for _, n := range table.nodeTable.nurseryNodes() {
			if n.ID().IsBlank() {
				node, err := table.call.register(n.addr)
				if err != nil {
					table.coreCtx.GetLog().Error("loadNurseryNodes", zap.Error(err))
					n.errorNum++
					return
				} else {
					if !node.ID().IsBlank() {
						table.addNode(node)
						n.SetID(node.ID())
						table.coreCtx.GetDB().GetAddressModel().UpdateServerNameByAddress(n.addr.String(), node.ServerName())
						deleteNodes = append(deleteNodes, n)
					}
				}
			}
		}
		table.nodeTable.removeNurseryNodes(deleteNodes)
	}
}

func (table *Table) doRefresh(done chan struct{}) {
	defer close(done)
	table.loadNurseryNodes()
	table.lookup()
}
func (table *Table) lookup() {
	table.lookupSelf()
	for i := 0; i < 3; i++ {
		table.lookupRand()
	}
}
func (table *Table) lookupByTarget(target ID) {
	nodes := table.nodeTable.queryNodesByIdAndDistance(target, findnodeResultLimit)
	for _, n := range nodes.entries {
		nodes, err := table.call.findNode(target, n)
		if err != nil {
			table.coreCtx.GetLog().Error("lookupByTarget", zap.Error(err))
			n.errorNum++
			return
		}
		for _, n2 := range nodes {
			table.addNode(n2)
		}
	}
}

func (table *Table) lookupSelf() {
	id := table.localNode.ID()
	table.lookupByTarget(id)
}

func (table *Table) lookupRand() {
	var target ID
	table.rand.Read(target[:])
	table.lookupByTarget(target)
}

func (table *Table) validate(node *Node, index int) {
	value, err := table.call.register(node.addr)
	if err == nil {
		if node.id != value.id {
			table.nodeTable.deleteNode(node)
			table.coreCtx.GetDB().GetAddressModel().UpdateServerNameByAddress(node.addr.String(), node.ServerName())
			table.addNode(value)
			return
		}
		node.liveNessChecks++
		table.nodeTable.bumpInBucket(node)
		return
	} else {
		table.coreCtx.GetLog().Error("validate", zap.Error(err))
	}

	node.errorNum++
	table.nodeTable.replace(index, node)
}

func (table *Table) doRevalidate(done chan struct{}) {
	defer close(done)
	table.loadNurseryNodes()
	node, _, index := table.nodeTable.nodeToRevalidate()
	if node != nil {
		table.validate(node, index)
	}

}
func NewTable(coreCtx *core.Context, localNode *Node, call *call) *Table {
	return &Table{rand: rand.New(rand.NewSource(0)), coreCtx: coreCtx, nodeTable: NewNodeTable(localNode, coreCtx), localNode: localNode, call: call}
}
