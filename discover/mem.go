package discover

import (
	"github.com/chuccp/shareExplorer/util"
	"math/rand"
	"net"
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

type NodeTable struct {
	buckets    [nBuckets]*bucket
	nursery    []*Node //bootstrap nodes
	serverNode map[ID]*Node
	localNode  *Node
	ips        DistinctNetSet //IP计数
	rand       *rand.Rand
}

func NewNodeTable(localNode *Node) *NodeTable {
	tab := &NodeTable{localNode: localNode, serverNode: make(map[ID]*Node), rand: rand.New(rand.NewSource(0))}
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

func (nodeTable *NodeTable) replace(index int, last *Node) {
	b := nodeTable.bucket(last.ID())
	if len(b.replacements) == 0 {
		b.entries = deleteNode0(b.entries, last)
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
	key := n.ID()
	k, fa := nodeTable.serverNode[key]
	if fa {
		k.lastUpdateTime = time.Now()
	} else {
		n.lastUpdateTime = time.Now()
		n.addTime = time.Now()
		nodeTable.serverNode[key] = n
	}
}
func (nodeTable *NodeTable) queryServer(id ID) (*Node, bool) {
	n, ok := nodeTable.serverNode[id]
	if ok {
		return n, ok
	} else {
		return nil, false
	}
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
