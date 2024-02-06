package discover

import (
	"container/list"
	"context"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"sync"
	"time"
)

type nodeSearchManage struct {
	table        *Table
	nodeSearches []*nodeSearch
	coreCtx      *core.Context
}

func NewNodeSearchManage(table *Table) *nodeSearchManage {
	return &nodeSearchManage{table: table, nodeSearches: make([]*nodeSearch, 0)}
}

func (nsm *nodeSearchManage) FindNodeStatus(searchId ID, isStart bool) *entity.NodeStatus {
	for _, search := range nsm.nodeSearches {
		if searchId == search.searchNode.id {
			if isStart && !search.nodeStatus.IsComplete() {
				go search.tempRun()
			}
			return search.nodeStatus
		}
	}
	nodeSearch := newNodeSearch(nsm.table, searchId)
	nsm.nodeSearches = append(nsm.nodeSearches, nodeSearch)
	go nodeSearch.run()
	go nodeSearch.tempRun()
	return nodeSearch.nodeStatus
}
func (nsm *nodeSearchManage) stopAll() {
	for _, search := range nsm.nodeSearches {
		search.stop()
	}
}

func (nsm *nodeSearchManage) run() {
	for _, search := range nsm.nodeSearches {
		search.run()
	}
}

type findValueNode struct {
	queryNode *Node
	fromId    ID
}

type findValueNodeQueue struct {
	useNode  map[string]*findValueNode
	nodeList *list.List
}

func (f *findValueNodeQueue) addNode0(node *findValueNode) {
	servername := node.queryNode.ServerName()
	_, ok := f.useNode[servername]
	if ok {
		return
	}
	f.useNode[node.queryNode.ServerName()] = node
	f.nodeList.PushBack(node)
}
func (f *findValueNodeQueue) addNode(fromId ID, queryNode *Node) {

	fvn := &findValueNode{queryNode: queryNode, fromId: fromId}
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

type queryTable interface {
	ID() ID
	Ping(node *Node) error
	FindRemoteValue(target ID, node *Node, distances int) ([]*Node, error)
	FindValue(target ID, distances int) []*Node
}

type queryNode struct {
	ctxCancel  context.CancelFunc
	ctx        context.Context
	queryTable queryTable
	searchId   ID
	once       sync.Once
}

func NewQueryNode(queryTable queryTable, searchId ID, parentCtx context.Context) *queryNode {
	ctx, ctxCancel := context.WithCancel(parentCtx)
	return &queryNode{queryTable: queryTable, searchId: searchId, ctx: ctx, ctxCancel: ctxCancel}
}
func (qn *queryNode) ping(node *Node) error {
	return qn.queryTable.Ping(node)
}
func (qn *queryNode) findValue(searchId ID, fromId ID, queryNode *Node) ([]*Node, error) {
	maxDistance := LogDist(searchId, fromId)
	queryDistance := LogDist(queryNode.id, fromId)
	if queryDistance > maxDistance {
		maxDistance = LogDist(searchId, queryNode.id)
	}
	return qn.queryTable.FindRemoteValue(searchId, queryNode, maxDistance)
}

func (qn *queryNode) StartFindNode() (*Node, error) {
	var findValueNodeQueue = NewFindValueNodeQueue()
	queryNode := qn.queryTable.FindValue(qn.searchId, 0)
	for _, n := range queryNode {
		if n.id == qn.searchId {
			err := qn.ping(n)
			if err != nil {
				return nil, err
			}
			return n, nil
		}
		findValueNodeQueue.addNode(qn.queryTable.ID(), n)
	}
	for {
		select {
		case <-qn.ctx.Done():
			{
				break
			}
		default:
			{
				node, fa := findValueNodeQueue.getNode()
				if !fa {
					return nil, QueryNotFoundError
				}
				queryNode, err := qn.findValue(qn.searchId, node.fromId, node.queryNode)
				if err == nil {
					for _, qNode := range queryNode {
						if qNode.id == qn.searchId {
							err := qn.ping(qNode)
							if err != nil {
								return nil, err
							}
							return qNode, nil
						}
						findValueNodeQueue.addNode(node.queryNode.id, qNode)
					}
				}
			}
		}
	}
	return nil, QueryCloseError
}
func (qn *queryNode) stop() {
	qn.once.Do(func() {
		qn.ctxCancel()
	})
}

type nodeSearch struct {
	queryTable    queryTable
	searchNode    *Node
	nodeStatus    *entity.NodeStatus
	ctxCancel     context.CancelFunc
	ctx           context.Context
	tempQueryNode *queryNode
	once          sync.Once
}

func newNodeSearch(queryTable queryTable, searchId ID) *nodeSearch {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &nodeSearch{queryTable: queryTable, searchNode: &Node{id: searchId}, nodeStatus: entity.NewNodeStatus(), ctx: ctx, ctxCancel: ctxCancel}
}
func (nodeSearch *nodeSearch) run() {
	go nodeSearch.loop()
}

func (nodeSearch *nodeSearch) tempRun() {
	nodeSearch.tempClose()
	queryNode := NewQueryNode(nodeSearch.queryTable, nodeSearch.searchNode.id, nodeSearch.ctx)
	nodeSearch.queryNode0(queryNode)
}
func (nodeSearch *nodeSearch) tempClose() {
	if nodeSearch.tempQueryNode != nil {
		nodeSearch.tempQueryNode.stop()
	}
}

func (nodeSearch *nodeSearch) stop() {
	nodeSearch.once.Do(func() {
		nodeSearch.ctxCancel()
	})
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
					go nodeSearch.ping(nodeSearch.searchNode)
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

func (nodeSearch *nodeSearch) queryNode(done chan<- struct{}) {
	defer close(done)
	if nodeSearch.searchNode == nil {
		queryNode := NewQueryNode(nodeSearch.queryTable, nodeSearch.searchNode.id, nodeSearch.ctx)
		nodeSearch.queryNode0(queryNode)
	}
}
func (nodeSearch *nodeSearch) queryNode0(qn *queryNode) {
	node, err := qn.StartFindNode()
	if err == nil {
		nodeSearch.searchNode = node
		nodeSearch.nodeStatus.SearchComplete(node.addr)
	} else {
		nodeSearch.nodeStatus.SearchFail(err)
	}
}

func (nodeSearch *nodeSearch) ping(node *Node) {
	err := nodeSearch.queryTable.Ping(node)
	if err != nil {
		nodeSearch.nodeStatus.SearchFail(err)
	}
}
func (nodeSearch *nodeSearch) FindValue(target ID, node *Node, distances int) (queryNode []*Node, err error) {
	return nodeSearch.queryTable.FindRemoteValue(target, node, distances)
}
