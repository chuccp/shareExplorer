package discover

import (
	"container/list"
	"context"
	"errors"
	"github.com/chuccp/shareExplorer/core"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	nBuckets                    = 17
	bucketMinDistance           = 239
	bucketIPLimit, bucketSubnet = 2, 24 //
	bucketSize                  = 16
	maxReplacements             = 10
	findnodeResultLimit         = 16
	lookupRequestLimit          = 3
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
type Table struct {
	mutex      sync.Mutex
	config     *core.Config
	buckets    [nBuckets]*bucket
	nursery    []*node //bootstrap nodes
	coreCtx    *core.Context
	localNode  *Node
	ips        DistinctNetSet //IP计数
	natClients *NodeStore
	clients    *NodeStore
	rand       *rand.Rand
	call       *call
	ctx        context.Context
	ctxCancel  context.CancelFunc
}

func (table *Table) loop() {
	var (
		revalidate     = time.NewTimer(table.nextRevalidateTime())
		refresh        = time.NewTimer(table.nextRefreshTime())
		revalidateDone chan struct{}
		refreshDone    = make(chan struct{})
	)
	go table.doRefresh(refreshDone)
	for {
		select {

		case <-table.ctx.Done():
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

func (table *Table) doRevalidate(done chan<- struct{}) {
	defer close(done)
	table.register()
}

func (table *Table) run() {

	//table.coreCtx.GetDB().GetAddressModel()

	go table.loop()
}
func (table *Table) stop() {
	if table.ctxCancel != nil {
		table.ctxCancel()
	}
}

type NodeStore struct {
	members   *list.List
	memberMap map[string]*node
	mutex     sync.RWMutex
}

func NewNodeStore() *NodeStore {
	return &NodeStore{members: new(list.List), memberMap: make(map[string]*node)}
}

func (nss *NodeStore) queryNode(pageNo, pageSize int) ([]*node, int) {
	nss.mutex.RLock()
	defer nss.mutex.RUnlock()
	list := nss.members
	nodes := make([]*node, 0)
	start := (pageNo - 1) * pageSize
	num := 0
	for ele := list.Front(); ele != nil; ele = ele.Next() {
		num++
		if num > start {
			nodes = append(nodes, ele.Value.(*node))
			if len(nodes) >= pageSize {
				break
			}
		}
	}
	return nodes, list.Len()
}
func (nss *NodeStore) addNode(n *node) {
	nss.mutex.Lock()
	defer nss.mutex.Unlock()
	key := n.serverName
	v, ok := nss.memberMap[key]
	if !ok {
		n.addedAt = time.Now()
		nss.members.PushFront(n)
	} else {
		v.addedAt = time.Now()
		v.addr = n.addr
	}
}

func (nss *NodeStore) queryONode(serverName string) *node {
	nss.mutex.RLock()
	defer nss.mutex.RUnlock()
	v, ok := nss.memberMap[serverName]
	if ok {
		return v
	}
	return nil
}

func (table *Table) addNursery(addr *net.UDPAddr) {
	if !containsAddress(table.nursery, addr) {
		n := wrapNode(NewNursery(addr))
		n.addedAt = time.Now()
		table.nursery = append(table.nursery, n)
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
func (table *Table) addReplacement(b *bucket, n *node) {
	for _, e := range b.replacements {
		if e.ID() == n.ID() {
			return // already in list
		}
	}
	if !table.addIP(b, n.IP()) {
		return
	}
	var removed *node
	b.replacements, removed = pushNode(b.replacements, n, maxReplacements)
	if removed != nil {
		table.removeIP(b, removed.IP())
	}
}
func (table *Table) removeIP(b *bucket, ip net.IP) {
	if IsLAN(ip) {
		return
	}
	table.ips.Remove(ip)
	b.ips.Remove(ip)
}
func pushNode(list []*node, n *node, max int) ([]*node, *node) {
	if len(list) < max {
		list = append(list, nil)
	}
	removed := list[len(list)-1]
	copy(list[1:], list)
	list[0] = n
	return list, removed
}

func (table *Table) collectTableNodes(rip net.IP, distances []uint, limit int) []*Node {
	var nodes []*Node
	var processed = make(map[uint]struct{})
	for _, dist := range distances {
		_, seen := processed[dist]
		if seen || dist > 256 {
			continue
		}

		// Get the nodes.
		var bn []*Node
		if dist == 0 {
			bn = []*Node{table.self()}
		} else if dist <= 256 {

			bn = unwrapNodes(table.bucketAtDistance(int(dist)).entries)

		}
		processed[dist] = struct{}{}
		for _, n := range bn {
			if CheckRelayIP(rip, n.IP()) != nil {
				continue
			}
			nodes = append(nodes, n)
			if len(nodes) >= limit {
				return nodes
			}
		}
	}
	return nodes
}

func (table *Table) HandleFindNode(rip net.IP, findNode *FindNode) []*Node {

	return table.collectTableNodes(rip, findNode.Distances, findnodeResultLimit)
}
func (table *Table) addIP(b *bucket, ip net.IP) bool {
	if len(ip) == 0 {
		return false // Nodes without IP cannot be added.
	}
	if IsLAN(ip) {
		return true
	}
	if !table.ips.Add(ip) {
		return false
	}
	if !b.ips.Add(ip) {
		table.ips.Remove(ip)
		return false
	}
	return true
}
func (table *Table) addNatClient(n *node) {
	table.natClients.addNode(n)
}
func (table *Table) addClient(n *node) {
	table.clients.addNode(n)
}
func (table *Table) addSeenNode(n *node) {
	if n.ID() == table.self().id {
		return
	}
	if !n.IsServer() {
		table.addClient(n)
		return
	}
	if n.IsNatClient() {
		table.addNatClient(n)
	}
	if n.IsNatServer() {
		b := table.bucket(n.ID())
		if contains(b.entries, n.ID()) {
			return
		}
		if len(b.entries) >= bucketSize {
			table.addReplacement(b, n)
			return
		}
		if !table.addIP(b, n.IP()) {
			return
		}
		n.addedAt = time.Now()
		b.entries = append(b.entries, n)
		b.replacements = deleteNode(b.replacements, n)
	}
}
func deleteNode(list []*node, n *node) []*node {
	for i := range list {
		if list[i].ID() == n.ID() {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}
func (table *Table) doRefresh(done chan<- struct{}) {
	defer close(done)
	table.loadSeedNodes()
	table.lookupSelf()
	for i := 0; i < 3; i++ {
		table.lookupRandom()
	}
}
func (table *Table) lookupSelf() {
	table.newLookup(table.coreCtx, table.self().ID())
}
func (table *Table) lookupRandom() {
	table.newLookup(table.coreCtx, table.self().ID())
}

func (table *Table) newLookup(ctx *core.Context, target ID) {
	newLookup(table, target, ctx, func(n *node) ([]*node, error) {
		return table.lookupWorker(n, target)
	}).run()
}
func (table *Table) newRandomLookup(ctx *core.Context) {
	var target ID
	table.rand.Read(target[:])
	table.newLookup(ctx, target)
}
func lookupDistances(target, dest ID) (dists []uint) {
	td := LogDist(target, dest)
	dists = append(dists, uint(td))
	for i := 1; len(dists) < lookupRequestLimit; i++ {
		if td+i <= 256 {
			dists = append(dists, uint(td+i))
		}
		if td-i > 0 {
			dists = append(dists, uint(td-i))
		}
	}
	return dists
}
func (table *Table) lookupWorker(destNode *node, target ID) ([]*node, error) {
	var (
		dists = lookupDistances(target, destNode.ID())
		nodes = nodesByDistance{target: target}
		err   error
	)
	var r []*Node
	r, err = table.findNode(unwrapNode(destNode), dists)
	if errors.Is(err, net.ErrClosed) {
		return nil, err
	}
	for _, n := range r {
		if n.ID() != table.self().ID() {
			nodes.push(wrapNode(n), findnodeResultLimit)
		}
	}
	return nodes.entries, err
}

func (table *Table) loadSeedNodes() {
	table.registerNursery()
}

func contains(ns []*node, id ID) bool {
	for _, n := range ns {
		if n.ID() == id {
			return true
		}
	}
	return false
}

func (table *Table) bucket(id ID) *bucket {
	d := LogDist(table.localNode.id, id)
	return table.bucketAtDistance(d)
}
func (table *Table) bucketAtDistance(d int) *bucket {
	if d <= bucketMinDistance {
		return table.buckets[0]
	}
	return table.buckets[d-bucketMinDistance-1]
}
func (table *Table) nodeToRevalidate() (n *node, bi int) {
	for _, bi = range table.rand.Perm(len(table.buckets)) {
		b := table.buckets[bi]
		if len(b.entries) > 0 {
			last := b.entries[len(b.entries)-1]
			return last, bi
		}
	}
	return nil, 0
}

func (table *Table) registerNursery() {
	for _, n := range table.nursery {
		if n.ID().IsBlank() {
			table.register0(n)
		} else {
			table.addSeenNode(n)
		}
	}
}

func (table *Table) register() {
	node, _ := table.nodeToRevalidate()
	if node != nil {
		table.register0(node)
	}

}

func (table *Table) register0(node *node) {
	value, err := table.call.register(table.localNode, node.addr.String())
	if err != nil {
		log.Println(err)
		return
	}
	node.liveNessChecks++
	node.SetID(value.ID())
}

func (table *Table) findNode(n *Node, distances []uint) ([]*Node, error) {
	nodes, err := table.call.findNode(table.localNode, n, n.addr.String(), distances)
	return nodes, err

}
func (table *Table) findNodeByID(target ID, nResults int, preferLive bool) *nodesByDistance {
	nodes := &nodesByDistance{target: target}
	liveNodes := &nodesByDistance{target: target}
	for _, b := range &table.buckets {
		for _, n := range b.entries {
			nodes.push(n, nResults)
			if preferLive && n.liveNessChecks > 0 {
				liveNodes.push(n, nResults)
			}
		}
	}
	if preferLive && len(liveNodes.entries) > 0 {
		return liveNodes
	}
	return nodes
}
func (table *Table) self() *Node {
	return table.localNode
}

func (table *Table) queryNode(nodeType, pageNo, pageSize int) ([]*node, int) {
	var nodes []*node
	if nodeType == client {
		return table.clients.queryNode(pageNo, pageSize)
	}

	return nodes, 0
}

func (table *Table) queryOneNode(serverName string) *node {
	node := table.natClients.queryONode(serverName)
	return node
}

func (table *Table) FindLocalNodeAddress() (string, error) {
	serverName := table.localNode.serverName
	n, err := table.FindValue(serverName)
	if err != nil {
		return "", err
	}
	return n.addr.String(), nil
}
func (table *Table) FindValue(serverName string) (*node, error) {

	table.call.queryNode(table.localNode)

	return nil, nil
}

func NewTable(coreCtx *core.Context, localNode *Node, call *call) *Table {

	ctx, ctxCancel := context.WithCancel(context.Background())

	table := &Table{
		coreCtx:    coreCtx,
		localNode:  localNode,
		rand:       rand.New(rand.NewSource(0)),
		call:       call,
		natClients: NewNodeStore(),
		clients:    NewNodeStore(),
		ctx:        ctx,
		ctxCancel:  ctxCancel,
	}
	for i := range table.buckets {
		table.buckets[i] = &bucket{
			index: i,
			ips:   DistinctNetSet{Subnet: bucketSubnet, Limit: bucketIPLimit},
		}
	}
	return table
}
