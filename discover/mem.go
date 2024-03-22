package discover

import (
	"container/list"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/util"
	"go.uber.org/zap"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	nBuckets                    = 17
	bucketMinDistance           = 239
	bucketIPLimit, bucketSubnet = 2, 24 //
	tableIPLimit, tableSubnet   = 10, 24
	bucketSize                  = 16
	maxReplacements             = 10
	findnodeResultLimit         = 16
	findValueResultLimit        = 34
	lookupRequestLimit          = 3
)

type bucket struct {
	entries      []*Node
	replacements []*Node
	index        int
	ips          DistinctNetSet
}
type mNode struct {
	n   *Node
	ele *list.Element
}
type nodeStore struct {
	nodeMap  map[ID]*mNode
	nodeList *list.List
	mutex    sync.Mutex
}

func (ns *nodeStore) add(node *Node) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	n, ok := ns.nodeMap[node.id]
	if ok {
		node.lastUpdateTime = time.Now()
		node.addTime = n.n.addTime
		ns.nodeMap[node.id].n = node
	} else {
		node.lastUpdateTime = time.Now()
		node.addTime = time.Now()
		ele := ns.nodeList.PushBack(node.id)
		ns.nodeMap[node.id] = &mNode{n: node, ele: ele}
	}
}
func (ns *nodeStore) remove(id ID) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	delete(ns.nodeMap, id)
	n, ok := ns.nodeMap[id]
	if ok {
		ns.nodeList.Remove(n.ele)
		delete(ns.nodeMap, id)
	}
}
func (ns *nodeStore) get(id ID) (*Node, bool) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	v, ok := ns.nodeMap[id]
	return v.n, ok
}
func newNodeStore() *nodeStore {
	return &nodeStore{nodeMap: make(map[ID]*mNode), nodeList: list.New()}
}

type NodeTable struct {
	buckets    [nBuckets]*bucket
	nursery    []*Node //bootstrap nodes
	serverNode *nodeStore
	localNode  *Node
	ips        DistinctNetSet //IP计数
	rand       *rand.Rand
	coreCtx    *core.Context
}

func NewNodeTable(localNode *Node, coreCtx *core.Context) *NodeTable {
	tab := &NodeTable{coreCtx: coreCtx, localNode: localNode, serverNode: newNodeStore(), rand: rand.New(rand.NewSource(0))}
	for i := range tab.buckets {
		tab.buckets[i] = &bucket{
			index: i,
			ips:   DistinctNetSet{Subnet: bucketSubnet, Limit: bucketIPLimit},
		}
	}
	return tab
}
func (nodeTable *NodeTable) nurseryNodes() []*Node {
	return nodeTable.nursery
}
func (nodeTable *NodeTable) hasNurseryNodes() bool {
	return len(nodeTable.nursery) > 0
}

func (nodeTable *NodeTable) queryNatServerForPage(pageNo, pageSize int) []*Node {
	if pageNo < 1 {
		pageNo = 1
	}
	nodes := make([]*Node, 0)
	keep := (pageNo - 1) * pageSize
	for _, b := range &nodeTable.buckets {
		for _, n := range b.entries {
			if keep == 0 {
				nodes = append(nodes, n)
				if len(nodes) == pageSize {
					return nodes
				}
			} else {
				keep--
			}
		}
	}
	return nodes
}

func (nodeTable *NodeTable) queryNodesByIdAndDistance(target ID, maxNum int) *nodesByDistance {
	nodes := &nodesByDistance{target: target}
	for _, b := range &nodeTable.buckets {
		for _, n := range b.entries {
			nodes.push(n, maxNum)
		}
	}
	return nodes
}

func (nodeTable *NodeTable) deleteNode(n *Node) {
	b := nodeTable.bucket(n.ID())
	b.entries = deleteNode0(b.entries, n)
}

func (nodeTable *NodeTable) nodeToRevalidate() (n *Node, bi int, index int) {

	for _, bi = range nodeTable.rand.Perm(len(nodeTable.buckets)) {
		b := nodeTable.buckets[bi]
		if len(b.entries) > 0 {
			nodeTable.coreCtx.GetLog().Debug("nodeToRevalidate", zap.Int("bucketIndex", bi))
			index = len(b.entries) - 1
			last := b.entries[index]
			return last, bi, index
		}
	}
	return nil, 0, 0
}

func (nodeTable *NodeTable) collectTableNodes(rip net.IP, distances []uint, limit int) []*Node {
	var nodes []*Node
	var processed = make(map[uint]struct{})
	for _, dist := range distances {
		_, seen := processed[dist]
		if seen || dist > 256 {
			continue
		}
		var bn []*Node
		if dist == 0 {
			bn = []*Node{nodeTable.localNode}
		} else if dist <= 256 {
			bn = nodeTable.bucketAtDistance(int(dist)).entries
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

type RecordBuckets struct {
	entries  []*Node
	maxElems int
}

func (recordBuckets *RecordBuckets) push(node *Node) bool {
	recordBuckets.entries = append(recordBuckets.entries, node)
	if len(recordBuckets.entries) >= recordBuckets.maxElems {
		return true
	}
	return false
}

func (nodeTable *NodeTable) collectTableFindNode(distances int) (queryNode []*Node) {
	var recordBuckets = &RecordBuckets{maxElems: findValueResultLimit}
	nodeTable.collectBucketsFindNode(0, nodeTable.bucketIndexAtDistance(distances), recordBuckets)
	nodeTable.collectBucketsFindNode(nodeTable.bucketIndexAtDistance(distances), nBuckets, recordBuckets)
	return recordBuckets.entries
}

func (nodeTable *NodeTable) collectBucketsFindNode(minBucketIndex, maxBucketIndex int, recordBuckets *RecordBuckets) {
	index := 0
	for {
		fa := nodeTable.collectBucketsByIndex(index, minBucketIndex, maxBucketIndex, recordBuckets)

		if fa {
			break
		}
		if len(recordBuckets.entries) >= recordBuckets.maxElems {
			break
		}
		index++
	}
}
func (nodeTable *NodeTable) collectBucketsByIndex(index, minBucketIndex, maxBucketIndex int, recordBuckets *RecordBuckets) bool {
	isEnd := true
	for i := minBucketIndex; i < maxBucketIndex; i++ {
		b := nodeTable.buckets[i]
		if len(b.entries) > index {
			if isEnd {
				isEnd = false
			}
			fill := recordBuckets.push(b.entries[index])
			if fill {
				return true
			}
		}
	}
	return false
}
func (nodeTable *NodeTable) bumpInBucket(last *Node) bool {
	b := nodeTable.bucket(last.ID())
	for i := range b.entries {
		if b.entries[i].ID() == last.ID() {
			if !last.IP().Equal(b.entries[i].IP()) {
				// Endpoint has changed, ensure that the new IP fits into table limits.
				nodeTable.removeIP(b, b.entries[i].IP())
				if !nodeTable.addIP(b, last.IP()) {
					// It doesn't, put the previous one back.
					nodeTable.addIP(b, b.entries[i].IP())
					return false
				}
			}
			// Move it to the front.
			copy(b.entries[1:], b.entries[:i])
			b.entries[0] = last
			return true
		}
	}
	return false
}
func (nodeTable *NodeTable) replace(index int, last *Node) {
	b := nodeTable.bucket(last.ID())
	if len(b.entries) == 0 || b.entries[len(b.entries)-1].ID() != last.ID() {
		return
	}
	if len(b.replacements) == 0 {
		nodeTable.coreCtx.GetLog().Debug("replace", zap.Int("replacements", len(b.replacements)), zap.Int("b.entries", len(b.entries)))
		b.entries = deleteNode0(b.entries, last)
		nodeTable.coreCtx.GetLog().Debug("replace", zap.Int("replacements", len(b.replacements)), zap.Int("b.entries", len(b.entries)))
		return
	}
	r := b.replacements[nodeTable.rand.Intn(len(b.replacements))]
	b.replacements = deleteNode0(b.replacements, r)
	b.entries[index] = r
}

func (nodeTable *NodeTable) removeNurseryNodes(ns []*Node) {
	if len(ns) > 0 {
		nodeTable.nursery = util.DeleteElements(nodeTable.nursery, ns)
	}
}
func (nodeTable *NodeTable) addNode(n *Node) {
	if n.IsNatServer() {
		nodeTable.addNatServer(n)
	}
	if n.IsServer() {
		nodeTable.addServer(n)
	}
}
func (nodeTable *NodeTable) bucketIndexAtDistance(d int) int {
	if d <= bucketMinDistance {
		return 0
	}
	return d - bucketMinDistance - 1
}
func (nodeTable *NodeTable) bucketAtDistance(d int) *bucket {
	return nodeTable.buckets[nodeTable.bucketIndexAtDistance(d)]
}

func (nodeTable *NodeTable) bucket(id ID) *bucket {
	d := LogDist(nodeTable.localNode.id, id)
	return nodeTable.bucketAtDistance(d)
}

func nodesContainsId(ns []*Node, id ID) (*Node, bool) {
	for _, n := range ns {
		if n.ID() == id {
			return n, true
		}
	}
	return nil, false
}
func (nodeTable *NodeTable) addReplacement(b *bucket, n *Node) {
	for _, e := range b.replacements {
		if e.ID() == n.ID() {
			return // already in list
		}
	}
	if !nodeTable.addIP(b, n.IP()) {
		return
	}
	var removed *Node
	b.replacements, removed = pushNode0(b.replacements, n, maxReplacements)
	if removed != nil {
		nodeTable.removeIP(b, removed.IP())
	}
}
func (nodeTable *NodeTable) removeIP(b *bucket, ip net.IP) {
	if IsLAN(ip) {
		return
	}
	nodeTable.ips.Remove(ip)
	b.ips.Remove(ip)
}
func pushNode0(list []*Node, n *Node, max int) ([]*Node, *Node) {
	if len(list) < max {
		list = append(list, nil)
	}
	removed := list[len(list)-1]
	copy(list[1:], list)
	list[0] = n
	return list, removed
}

func (nodeTable *NodeTable) addIP(b *bucket, ip net.IP) bool {
	if len(ip) == 0 {
		return false // Nodes without IP cannot be added.
	}
	if IsLAN(ip) {
		return true
	}
	if !nodeTable.ips.Add(ip) {
		return false
	}
	if !b.ips.Add(ip) {
		nodeTable.ips.Remove(ip)
		return false
	}
	return true
}

func (nodeTable *NodeTable) addNatServer(n *Node) {
	b := nodeTable.bucket(n.ID())
	preNode, fa := nodesContainsId(b.entries, n.ID())
	if fa {
		preNode.lastUpdateTime = time.Now()
		return
	}
	if len(b.entries) >= bucketSize {
		nodeTable.addReplacement(b, n)
		return
	}
	if !nodeTable.addIP(b, n.IP()) {
		return
	}
	n.addTime = time.Now()
	b.entries = append(b.entries, n)
	b.replacements = deleteNode0(b.replacements, n)
}
func (nodeTable *NodeTable) addServer(n *Node) {
	nodeTable.serverNode.add(n)

}
func (nodeTable *NodeTable) queryServer(id ID) (*Node, bool) {
	return nodeTable.serverNode.get(id)
}

func deleteNode0(list []*Node, n *Node) []*Node {
	for i := range list {
		if list[i].ID() == n.ID() {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}

func (nodeTable *NodeTable) addSeedNode(n *Node) {
	if !n.HasId() {
		nodeTable.addNursery(n)
	} else {
		nodeTable.addNode(n)
	}
}

func (nodeTable *NodeTable) addNursery(n *Node) {
	if !nodesContainsAddress(nodeTable.nursery, n.addr) {
		n.addTime = time.Now()
		nodeTable.nursery = append(nodeTable.nursery, n)
	}
}

func nodesContainsAddress(ns []*Node, addr *net.UDPAddr) bool {
	for _, v := range ns {
		if IsSameAddress(v.addr, addr) {
			return true
		}
	}
	return false
}
