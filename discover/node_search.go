package discover

import (
	"container/list"
	"context"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"log"
	"time"
)

type nodeSearchManage struct {
	table        *Table
	nodeSearches []*nodeSearch
	coreCtx      *core.Context
}

func NewNodeSearchManage(table *Table) *nodeSearchManage {
	return &nodeSearchManage{table: table}
}

func (nsm *nodeSearchManage) FindNodeStatus(searchId ID) *entity.NodeStatus {
	for _, search := range nsm.nodeSearches {
		if searchId == search.searchNode.id {
			return search.nodeStatus
		}
	}
	nodeSearch := newNodeSearch(nsm.table, searchId)
	nsm.nodeSearches = append(nsm.nodeSearches, nodeSearch)
	go nodeSearch.run()
	return nodeSearch.nodeStatus
}
func (nsm *nodeSearchManage) stop() {
	for _, search := range nsm.nodeSearches {
		search.stop()
	}
}

func (nsm *nodeSearchManage) run() {

	//nsm.table

}

type nodeSearch struct {
	table      *Table
	searchNode *Node
	nodeStatus *entity.NodeStatus
	ctxCancel  context.CancelFunc
	ctx        context.Context
}

func newNodeSearch(table *Table, searchId ID) *nodeSearch {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &nodeSearch{table: table, searchNode: &Node{id: searchId}, nodeStatus: entity.NewNodeStatus(), ctx: ctx, ctxCancel: ctxCancel}
}
func (nodeSearch *nodeSearch) run() {
	go nodeSearch.loop()
}
func (nodeSearch *nodeSearch) stop() {
	if nodeSearch.ctxCancel != nil {
		nodeSearch.ctxCancel()
	}

}
func (nodeSearch *nodeSearch) updateNodeStatus() {

}
func (nodeSearch *nodeSearch) loop() {
	var (
		queryNode     = time.NewTimer(time.Second * 10)
		ping          = time.NewTimer(time.Second * 10)
		pingDone      chan struct{}
		queryNodeDone = make(chan struct{})
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

		case <-ping.C:
			{

				if pingDone == nil {
					pingDone = make(chan struct{})
					go nodeSearch.queryNode(pingDone)
				}

			}
		case <-pingDone:
			{
				ping.Reset(time.Second * 10)
				pingDone = nil
			}
		case <-nodeSearch.ctx.Done():
			{
				break
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
	localId     ID
}

type findValueNodeQueue struct {
	useNode  map[string]*findValueNode
	nodeList *list.List
}

func (f *findValueNodeQueue) addNode0(node *findValueNode) {
	f.useNode[node.queryNode.ServerName()] = node
	f.nodeList.PushBack(node)
}
func (f *findValueNodeQueue) addNode(localId ID, preId ID, fromId ID, node *Node) {
	_, ok := f.useNode[node.ServerName()]
	if ok {
		return
	}
	maxDistance := LogDist(preId, fromId)
	queryDistance := LogDist(node.ID(), fromId)
	fvn := &findValueNode{maxDistance: queryDistance, queryNode: node, fromId: fromId}
	if queryDistance > maxDistance {
		fvn.maxDistance = LogDist(localId, node.ID())
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
func NewFindValueNodeQueue() *findValueNodeQueue {
	return &findValueNodeQueue{useNode: make(map[string]*findValueNode), nodeList: new(list.List)}
}

func (nodeSearch *nodeSearch) queryNode(done chan<- struct{}) {
	defer close(done)
	log.Println("queryNode")
	var findValueNodeQueue = NewFindValueNodeQueue()
	queryNode := nodeSearch.table.FindValue(nodeSearch.searchNode.ServerName(), 0)
	for _, n := range queryNode {
		log.Println("queryNode", n.id.String())
		if n.id == nodeSearch.searchNode.id {
			nodeSearch.ping(n)
			return
		}
		findValueNodeQueue.addNode(nodeSearch.searchNode.id, nodeSearch.searchNode.id, nodeSearch.table.localNode.id, n)
	}
	for {
		node, fa := findValueNodeQueue.getNode()
		if !fa {
			break
		}
		queryNode, err := nodeSearch.FindValue(nodeSearch.searchNode.ServerName(), node.queryNode, node.maxDistance)
		if err == nil {
			for _, qNode := range queryNode {
				if qNode.id == nodeSearch.searchNode.id {
					nodeSearch.ping(qNode)
					return
				}
				findValueNodeQueue.addNode(nodeSearch.searchNode.id, node.fromId, node.queryNode.id, qNode)
			}
		}
	}
}

func (nodeSearch *nodeSearch) ping(node *Node) {
	nodeSearch.searchNode = node
	nodeSearch.nodeStatus.SearchComplete(node.addr)
}

func (nodeSearch *nodeSearch) FindValue(target string, node *Node, distances int) (queryNode []*Node, err error) {
	return nodeSearch.table.call.findValue(target, distances, node.addr)
}
