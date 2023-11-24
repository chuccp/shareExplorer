package discover

import (
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
type TableStore struct {
	tableStore *sync.Map
	context    *core.Context
}

func NewTableStore(context *core.Context) *TableStore {
	return &TableStore{tableStore: new(sync.Map), context: context}
}
func (ts *TableStore) AddTable(localNode *Node) *Table {
	table := NewTable(ts.context, localNode)
	t, _ := ts.tableStore.LoadOrStore(localNode.id, table)
	return t.(*Table)
}
func (ts *TableStore) RangeTable(f func(key ID, value *Table) bool) {
	ts.tableStore.Range(func(key, value any) bool {
		return f(key.(ID), value.(*Table))
	})
}
func (ts *TableStore) GetTable() *Table {
	var t *Table
	ts.tableStore.Range(func(key, value any) bool {
		t = value.(*Table)
		return false
	})
	return t
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
func (tableGroup *TableGroup) GetOneTable() *Table {
	return tableGroup.tableStore.GetTable()
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
	if !n.IsServer() {
		return
	}
	tableGroup.mutex.Lock()
	defer tableGroup.mutex.Unlock()
	tableGroup.tableStore.RangeTable(func(key ID, value *Table) bool {
		if n.id != key {
			value.addSeenNode(n)
		}
		return true
	})
}

func (tableGroup *TableGroup) AddTable(localNode *Node) *Table {
	return tableGroup.tableStore.AddTable(localNode)
}

type NatClientStore struct {
	members *sync.Map
	num     uint32
}

func NewNatServerStore() *NatClientStore {
	return &NatClientStore{members: new(sync.Map)}
}

func (nss *NatClientStore) addNode(n *node) {
	n.addedAt = time.Now()
	nss.members.Store(n.serverName, n)
}

type Table struct {
	buckets   [nBuckets]*bucket
	nursery   []*node //bootstrap nodes
	context   *core.Context
	localNode *Node
	ips       DistinctNetSet //IP计数
	clients   *NatClientStore
	rand      *rand.Rand
	call      *call
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
func (tab *Table) addReplacement(b *bucket, n *node) {
	for _, e := range b.replacements {
		if e.ID() == n.ID() {
			return // already in list
		}
	}
	if !tab.addIP(b, n.IP()) {
		return
	}
	var removed *node
	b.replacements, removed = pushNode(b.replacements, n, maxReplacements)
	if removed != nil {
		tab.removeIP(b, removed.IP())
	}
}
func (tab *Table) removeIP(b *bucket, ip net.IP) {
	if IsLAN(ip) {
		return
	}
	tab.ips.Remove(ip)
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

func (tab *Table) collectTableNodes(rip net.IP, distances []uint, limit int) []*Node {
	var nodes []*Node
	var processed = make(map[uint]struct{})
	for _, dist := range distances {
		// Reject duplicate / invalid distances.
		_, seen := processed[dist]
		if seen || dist > 256 {
			continue
		}

		// Get the nodes.
		var bn []*Node
		if dist == 0 {
			bn = []*Node{tab.self()}
		} else if dist <= 256 {

			bn = unwrapNodes(tab.bucketAtDistance(int(dist)).entries)

		}
		processed[dist] = struct{}{}

		// Apply some pre-checks to avoid sending invalid nodes.
		for _, n := range bn {
			// TODO livenessChecks > 1
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

func (tab *Table) HandleFindNode(rip net.IP, findNode *FindNode) []*Node {

	return tab.collectTableNodes(rip, findNode.Distances, findnodeResultLimit)
}
func (tab *Table) addIP(b *bucket, ip net.IP) bool {
	if len(ip) == 0 {
		return false // Nodes without IP cannot be added.
	}
	if IsLAN(ip) {
		return true
	}
	if !tab.ips.Add(ip) {
		return false
	}
	if !b.ips.Add(ip) {
		tab.ips.Remove(ip)
		return false
	}
	return true
}
func (tab *Table) addClient(n *node) {
	tab.clients.addNode(n)
}
func (tab *Table) addSeenNode(n *node) {
	if n.ID() == tab.self().id {
		return
	}
	if n.IsNatClient() {
		tab.addClient(n)
	}
	if n.IsNatServer() {
		b := tab.bucket(n.ID())
		if contains(b.entries, n.ID()) {
			return
		}
		if len(b.entries) >= bucketSize {
			tab.addReplacement(b, n)
			return
		}
		if !tab.addIP(b, n.IP()) {
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
func (tab *Table) doRefresh() {
	tab.loadSeedNodes()
	tab.lookupSelf()
	for i := 0; i < 3; i++ {
		tab.lookupRandom()
	}
}
func (tab *Table) lookupSelf() {
	tab.newLookup(tab.context, tab.self().ID())
}
func (tab *Table) lookupRandom() {
	tab.newLookup(tab.context, tab.self().ID())
}

func (tab *Table) newLookup(ctx *core.Context, target ID) {
	newLookup(tab, target, ctx, func(n *node) ([]*node, error) {
		return tab.lookupWorker(n, target)
	}).run()
}
func (tab *Table) newRandomLookup(ctx *core.Context) {
	var target ID
	tab.rand.Read(target[:])
	tab.newLookup(ctx, target)
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
func (tab *Table) lookupWorker(destNode *node, target ID) ([]*node, error) {
	var (
		dists = lookupDistances(target, destNode.ID())
		nodes = nodesByDistance{target: target}
		err   error
	)
	var r []*Node
	r, err = tab.findNode(unwrapNode(destNode), dists)
	if errors.Is(err, net.ErrClosed) {
		return nil, err
	}
	for _, n := range r {
		if n.ID() != tab.self().ID() {
			nodes.push(wrapNode(n), findnodeResultLimit)
		}
	}
	return nodes.entries, err
}

func (tab *Table) loadSeedNodes() {
	tab.registerNursery()
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
		if n.ID().IsBlank() {
			tab.register0(n)
		} else {
			tab.addSeenNode(n)
		}
	}
}

func (tab *Table) register() {
	node, _ := tab.nodeToRevalidate()
	if node != nil {
		tab.register0(node)
	}

}

func (tab *Table) register0(node *node) {
	value, err := tab.call.register(tab.localNode, node.addr.String())
	if err != nil {
		log.Println(err)
		return
	}
	node.liveNessChecks++
	node.SetID(value.ID())
}

func (tab *Table) findNode(n *Node, distances []uint) ([]*Node, error) {
	nodes, err := tab.call.findNode(tab.localNode, n, n.addr.String(), distances)
	return nodes, err

}
func (tab *Table) findnodeByID(target ID, nresults int, preferLive bool) *nodesByDistance {
	nodes := &nodesByDistance{target: target}
	liveNodes := &nodesByDistance{target: target}
	for _, b := range &tab.buckets {
		for _, n := range b.entries {
			nodes.push(n, nresults)
			if preferLive && n.liveNessChecks > 0 {
				liveNodes.push(n, nresults)
			}
		}
	}
	if preferLive && len(liveNodes.entries) > 0 {
		return liveNodes
	}
	return nodes
}
func (tab *Table) self() *Node {
	return tab.localNode
}

func NewTable(context *core.Context, localNode *Node) *Table {
	table := &Table{
		context:   context,
		localNode: localNode,
		rand:      rand.New(rand.NewSource(0)),
		call:      &call{httpClient: core.NewHttpClient(context)},
		clients:   NewNatServerStore(),
	}
	for i := range table.buckets {
		table.buckets[i] = &bucket{
			index: i,
			ips:   DistinctNetSet{Subnet: bucketSubnet, Limit: bucketIPLimit},
		}
	}
	return table
}
