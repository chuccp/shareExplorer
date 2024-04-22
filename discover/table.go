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

const (
	revalidateTime    = 10
	registerTime      = 120
	refreshTime       = 1800
	clearTime         = registerTime
	serverTimeoutTime = registerTime * 4
)

type Table struct {
	rand              *rand.Rand
	mutex             sync.Mutex
	nodeTable         *NodeTable
	localNode         *Node
	call              *call
	coreCtx           *core.Context
	ctx               context.Context
	ctxCancel         context.CancelFunc
	revalidateTime    time.Duration
	registerTime      time.Duration
	refreshTime       time.Duration
	clearTime         time.Duration
	serverTimeoutTime time.Duration
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

func (table *Table) AddNode(n *Node) {
	table.nodeTable.addNode(n)
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
		table.AddNode(seed)
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
func (table *Table) queryServerForPage(pageNo, pageSize int) ([]*Node, int) {
	return table.nodeTable.queryServerForPage(pageNo, pageSize)
}

func (table *Table) self() *Node {
	return table.localNode
}

func (table *Table) run() {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	if table.ctxCancel != nil {
		table.ctxCancel()
		table.ctxCancel = nil
	}
	table.ctx, table.ctxCancel = context.WithCancel(context.Background())
	err := table.loadAddress()
	if err != nil {
		return
	} else {

		table.coreCtx.Go(func() {
			table.loop()
		})
		table.coreCtx.Go(func() {
			table.loopRegister()
		})
	}
}

func (table *Table) loop() {
	var (
		revalidate  = time.NewTimer(table.nextRevalidateTime())
		refresh     = time.NewTimer(table.nextRefreshTime())
		clearServer = time.NewTimer(table.nextClearTime())

		revalidateDone  chan struct{}
		clearServerDone chan struct{}
		refreshDone     = make(chan struct{})
	)
	table.coreCtx.Go(func() {
		table.doRefresh(refreshDone)
	})
loop:
	for {
		select {

		case <-table.ctx.Done():
			{
				break loop
			}
		case <-refresh.C:
			{
				if refreshDone == nil {
					refreshDone = make(chan struct{})
					table.coreCtx.Go(func() {
						table.doRefresh(refreshDone)
					})
				}

			}
		case <-revalidate.C:
			{
				if revalidateDone == nil {
					revalidateDone = make(chan struct{})
					table.coreCtx.Go(func() {
						table.doRevalidate(revalidateDone)
					})
				}

			}

		case <-clearServer.C:
			{
				if clearServerDone == nil {
					clearServerDone = make(chan struct{})
					table.coreCtx.Go(func() {
						table.doClearServer(clearServerDone)
					})
				}

			}
		case <-clearServerDone:
			{
				clearServer.Reset(table.nextClearTime())
				clearServerDone = nil
			}

		case <-revalidateDone:
			{
				revalidate.Reset(table.nextRevalidateTime())
				revalidateDone = nil
			}
		case <-refreshDone:
			refresh.Reset(table.nextRefreshTime())
			refreshDone = nil
		}
	}

	if revalidateDone != nil {
		<-revalidateDone
	}
	if refreshDone != nil {
		<-refreshDone
	}
	if clearServerDone != nil {
		<-clearServerDone
	}
	table.coreCtx.GetLog().Debug("close table")
}
func (table *Table) doClearServer(done chan struct{}) {
	defer close(done)
	table.coreCtx.GetLog().Debug("doClearServer", zap.Bool("isServer", table.self().isServer))
	table.nodeTable.clearServerTimeOut(table.serverTimeoutTime)
}
func (table *Table) loopRegister() {
	registerDone := make(chan struct{})
	register := time.NewTimer(table.nextRegisterTime())
	for {

		select {
		case <-register.C:
			{
				table.coreCtx.Go(func() {
					table.doRegister(registerDone)
				})
			}

		case <-registerDone:
			{
				register.Reset(table.nextRegisterTime())
				registerDone = make(chan struct{})
			}
		case <-table.ctx.Done():
			{
				return
			}
		}
	}

}

func (table *Table) isDone() bool {
	select {
	case <-table.ctx.Done():
		return true
	default:
		return false
	}
}

func (table *Table) doRegister(done chan struct{}) {
	close(done)
	table.coreCtx.GetLog().Debug("doRegister", zap.Bool("isServer", table.self().isServer))
	if table.self().isServer {
		queryNodes := table.nodeTable.collectLocalTableFindNode()
		for _, node := range queryNodes {
			if table.isDone() {
				return
			}
			table.coreCtx.GetLog().Debug("doRegister", zap.String("remoteAddress", node.addr.String()))
			table.validateNatServer(node)
		}
	}
}
func (table *Table) nextRegisterTime() time.Duration {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	half := table.registerTime / 2
	return half + time.Duration(table.rand.Int63n(int64(half)))
}

func (table *Table) nextClearTime() time.Duration {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	half := table.clearTime / 2
	return half + time.Duration(table.rand.Int63n(int64(half)))
}
func (table *Table) nextRevalidateTime() time.Duration {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	return time.Duration(table.rand.Int63n(int64(table.revalidateTime)))
}

func (table *Table) nextRefreshTime() time.Duration {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	half := table.refreshTime / 2
	return half + time.Duration(table.rand.Int63n(int64(half)))
}

func (table *Table) stop() {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	if table.ctxCancel != nil {
		table.coreCtx.GetLog().Debug("stop", zap.Bool("ctxCancel", true))
		table.ctxCancel()
		table.ctxCancel = nil
	}
}

func (table *Table) loadNurseryNodes() {
	if table.nodeTable.hasNurseryNodes() {
		deleteNodes := make([]*Node, 0)
		for _, n := range table.nodeTable.nurseryNodes() {
			if n.ID().IsBlank() {
				if table.isDone() {
					return
				}
				node, err := table.call.register(n.addr)
				if err != nil {
					table.coreCtx.GetLog().Error("loadNurseryNodes", zap.Error(err))
					n.errorNum++
					return
				} else {
					if !node.ID().IsBlank() {
						if node.id != table.localNode.id {
							table.AddNode(node)
							n.SetID(node.ID())
							table.coreCtx.GetDB().GetAddressModel().UpdateServerNameByAddress(n.addr.String(), node.ServerName())
						} else {
							table.coreCtx.GetLog().Info("loadNurseryNodes", zap.String("msg", "connection self"))
						}
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
	table.coreCtx.GetLog().Debug("doRefresh", zap.Bool("isServer", table.self().isServer))
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
		if table.isDone() {
			return
		}
		nodes, err := table.call.findNode(target, n)
		if err != nil {
			table.coreCtx.GetLog().Error("lookupByTarget", zap.Error(err))
			n.errorNum++
			continue
		}
		for _, n2 := range nodes {
			table.AddNode(n2)
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

func (table *Table) validateNatServer(node *Node) {
	value, err := table.call.register(node.addr)
	if err == nil {
		if node.id != value.id {
			table.nodeTable.deleteNode(node)
			table.coreCtx.GetDB().GetAddressModel().UpdateServerNameByAddress(node.addr.String(), node.ServerName())
			table.AddNode(value)
			return
		}
		node.liveNessChecks++
		table.nodeTable.bumpInBucket(node)
		return
	} else {
		table.coreCtx.GetLog().Error("validate", zap.Error(err))
	}
	node.errorNum++
	table.nodeTable.replace(node)
}

func (table *Table) doRevalidate(done chan struct{}) {
	defer close(done)
	table.loadNurseryNodes()
	node, _ := table.nodeTable.nodeToRevalidate()
	if node != nil {
		table.validateNatServer(node)
	}

}
func NewTable(coreCtx *core.Context, localNode *Node, call *call) *Table {
	table := &Table{rand: rand.New(rand.NewSource(0)), coreCtx: coreCtx, nodeTable: NewNodeTable(localNode, coreCtx), localNode: localNode, call: call}
	_revalidateTime_ := coreCtx.GetConfigInt64OrDefault("traversal", "revalidate.time", int64(revalidateTime))
	table.revalidateTime = time.Second * time.Duration(_revalidateTime_)

	_registerTime_ := coreCtx.GetConfigInt64OrDefault("traversal", "register.time", int64(registerTime))
	table.registerTime = time.Second * time.Duration(_registerTime_)

	_refreshTime_ := coreCtx.GetConfigInt64OrDefault("traversal", "refresh.time", int64(refreshTime))
	table.refreshTime = time.Second * time.Duration(_refreshTime_)

	_clearTime_ := coreCtx.GetConfigInt64OrDefault("traversal", "clear.time", int64(clearTime))
	table.clearTime = time.Second * time.Duration(_clearTime_)

	_serverTimeoutTime_ := coreCtx.GetConfigInt64OrDefault("traversal", "server.timeout.time", int64(serverTimeoutTime))
	table.serverTimeoutTime = time.Second * time.Duration(_serverTimeoutTime_)

	coreCtx.GetLog().Debug("NewTable", zap.Duration("revalidateTime", table.revalidateTime))
	coreCtx.GetLog().Debug("NewTable", zap.Duration("registerTime", table.registerTime))
	coreCtx.GetLog().Debug("NewTable", zap.Duration("refreshTime", table.refreshTime))
	coreCtx.GetLog().Debug("NewTable", zap.Duration("clearTime", table.clearTime))
	coreCtx.GetLog().Debug("NewTable", zap.Duration("serverTimeoutTime", table.serverTimeoutTime))

	return table
}
