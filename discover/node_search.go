package discover

import (
	"container/list"
	"context"
	"github.com/chuccp/shareExplorer/entity"
	"time"
)

type nodeSearch struct {
	table      *Table
	localNode  *Node
	remoteNode *Node
	nodeStatus *entity.NodeStatus
	ctxCancel  context.CancelFunc
	ctx        context.Context
}

func newNodeSearch(table *Table, localNode *Node) *nodeSearch {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &nodeSearch{table: table, localNode: localNode, nodeStatus: entity.NewNodeStatus(), ctx: ctx, ctxCancel: ctxCancel}
}
func (nodeSearch *nodeSearch) run() {
	go nodeSearch.loop()
}
func (nodeSearch *nodeSearch) stop() {

}
func (nodeSearch *nodeSearch) updateNodeStatus() {

}
func (nodeSearch *nodeSearch) loop() {
	var (
		queryNode     = time.NewTimer(time.Second * 10)
		refresh       = time.NewTimer(time.Second * 10)
		queryNodeDone chan struct{}
		refreshDone   = make(chan struct{})
	)
	go nodeSearch.queryNode(queryNodeDone)
	for {
		select {

		case <-queryNode.C:
			{

				if queryNodeDone == nil {
					queryNodeDone = make(chan struct{})
					go nodeSearch.queryNode(queryNodeDone)
				}

			}
		case <-queryNodeDone:
			{
				queryNode.Reset(time.Second * 10)
				queryNodeDone = nil
			}

		case <-refresh.C:
			{

				if refreshDone == nil {
					refreshDone = make(chan struct{})
					go nodeSearch.queryNode(refreshDone)
				}

			}
		case <-refreshDone:
			{
				refresh.Reset(time.Second * 10)
				refreshDone = nil
			}

		}
	}

}
func (nodeSearch *nodeSearch) refresh(done chan<- struct{}) {
	defer close(done)
}

type findValueNode struct {
	maxDistance int
	queryNode   *Node
	fromId      ID
}

type findValueNodeQueue struct {
	useNode   map[string]*findValueNode
	nodeList  *list.List
	localNode *Node
}

func (f *findValueNodeQueue) addNode0(node *findValueNode) {
	f.useNode[node.queryNode.serverName] = node
	f.nodeList.PushBack(node)
}
func (f *findValueNodeQueue) addNode(preId ID, fromId ID, node *Node) {
	if node.serverName == f.localNode.serverName {
		return
	}
	_, ok := f.useNode[node.serverName]
	if ok {
		return
	}
	maxDistance := LogDist(preId, fromId)
	queryDistance := LogDist(node.ID(), fromId)
	fvn := &findValueNode{maxDistance: queryDistance, queryNode: node, fromId: fromId}
	if queryDistance > maxDistance {
		fvn.maxDistance = LogDist(preId, node.ID())
	}
	f.addNode0(fvn)
}
func (f *findValueNodeQueue) getNode() (*findValueNode, bool) {
	ele := f.nodeList.Front()
	if ele != nil {
		return ele.Value.(*findValueNode), true
	}
	return nil, false
}
func NewFindValueNodeQueue(localNode *Node) *findValueNodeQueue {
	return &findValueNodeQueue{useNode: make(map[string]*findValueNode), nodeList: new(list.List), localNode: localNode}
}

func (nodeSearch *nodeSearch) queryNode(done chan<- struct{}) {
	defer close(done)
	var findValueNodeQueue = NewFindValueNodeQueue(nodeSearch.localNode)
	queryNode := nodeSearch.table.FindValue(nodeSearch.localNode.serverName, 0)
	for _, n := range queryNode {
		if n.serverName == nodeSearch.localNode.serverName {
			nodeSearch.ping(n)
			return
		}
		findValueNodeQueue.addNode(nodeSearch.localNode.id, nodeSearch.localNode.id, n)
	}
	for {
		node, fa := findValueNodeQueue.getNode()
		if !fa {
			break
		}
		queryNode, err := nodeSearch.FindValue(nodeSearch.localNode.serverName, node.queryNode, node.maxDistance)
		if err == nil {
			for _, qNode := range queryNode {
				if qNode.serverName == nodeSearch.localNode.serverName {
					nodeSearch.ping(qNode)
					return
				}
				findValueNodeQueue.addNode(node.fromId, node.queryNode.id, qNode)
			}
		}
	}
}

func (nodeSearch *nodeSearch) ping(node *Node) {
	nodeSearch.remoteNode = node
	nodeSearch.nodeStatus.SearchComplete(node.addr)
}

func (nodeSearch *nodeSearch) FindValue(target string, node *Node, distances int) (queryNode []*Node, err error) {
	return nodeSearch.table.call.findValue(target, distances, node.addr)
}
