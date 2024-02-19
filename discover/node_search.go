package discover

import (
	"container/list"
	"context"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"log"
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
	preId     ID
}

type findValueNodeQueue struct {
	useNode    map[ID]*findValueNode
	nodeList   *list.List
	queryTable queryTable
}

func (f *findValueNodeQueue) addNode0(node *findValueNode) {
	servername := node.queryNode.ID()
	_, ok := f.useNode[servername]
	if ok {
		return
	}
	f.useNode[node.queryNode.ID()] = node
	f.nodeList.PushBack(node)
}
func (f *findValueNodeQueue) addNode(preId ID, fromId ID, queryNode *Node) {
	maxDistance := LogDist(preId, fromId)
	queryDistance := LogDist(queryNode.id, fromId)
	if maxDistance == 0 || queryDistance < maxDistance {
		fvn := &findValueNode{queryNode: queryNode, fromId: fromId, preId: preId}
		f.addNode0(fvn)
	}
	f.queryTable.AddNatServer(wrapNode(queryNode))
}
func (f *findValueNodeQueue) getNode() (*findValueNode, bool) {
	ele := f.nodeList.Front()
	if ele != nil {
		fvn := ele.Value.(*findValueNode)
		f.nodeList.Remove(ele)
		return fvn, true
	}
	return nil, false
}
func NewFindValueNodeQueue(queryTable queryTable) *findValueNodeQueue {
	return &findValueNodeQueue{queryTable: queryTable, useNode: make(map[ID]*findValueNode), nodeList: new(list.List)}
}

type queryTable interface {
	ID() ID
	Ping(node *Node) error
	FindRemoteServer(target ID, node *Node, distances int) ([]*Node, error)
	FindServer(target ID, distances int) []*Node
	AddNatServer(n *node)
}

type queryValue struct {
	ctxCancel  context.CancelFunc
	ctx        context.Context
	queryTable queryTable
	searchId   ID
	once       sync.Once
}

func NewQueryValue(queryTable queryTable, searchId ID, parentCtx context.Context) *queryValue {
	ctx, ctxCancel := context.WithCancel(parentCtx)
	return &queryValue{queryTable: queryTable, searchId: searchId, ctx: ctx, ctxCancel: ctxCancel}
}
func (qv *queryValue) ping(node *Node) error {
	return qv.queryTable.Ping(node)
}
func (qv *queryValue) findValue(preId ID, fromId ID, searchId ID, queryNode *Node) ([]*Node, error) {
	maxDistance := LogDist(preId, fromId)
	queryDistance := LogDist(queryNode.id, fromId)
	if maxDistance != 0 && queryDistance >= maxDistance {
		queryDistance = LogDist(qv.queryTable.ID(), queryNode.id)
	}
	return qv.queryTable.FindRemoteServer(searchId, queryNode, queryDistance)
}

func (qv *queryValue) StartFindValue() (*Node, error) {
	var findValueNodeQueue = NewFindValueNodeQueue(qv.queryTable)
	queryNode := qv.queryTable.FindServer(qv.searchId, 0)
	for _, n := range queryNode {
		if n.id == qv.searchId {
			err := qv.ping(n)
			if err != nil {
				return nil, err
			}
			return n, nil
		}
		findValueNodeQueue.addNode(qv.queryTable.ID(), qv.queryTable.ID(), n)
	}
	for {
		select {
		case <-qv.ctx.Done():
			{
				break
			}
		default:
			{
				node, fa := findValueNodeQueue.getNode()
				if !fa {
					return nil, QueryNotFoundError
				}
				queryNodes, err := qv.findValue(node.preId, node.fromId, qv.searchId, node.queryNode)
				if err == nil {
					for _, qNode := range queryNodes {
						if qNode.id == qv.searchId {
							err := qv.ping(qNode)
							if err != nil {
								return nil, err
							}
							return qNode, nil
						}
						findValueNodeQueue.addNode(node.fromId, node.queryNode.id, qNode)
					}
				} else {
					log.Println(err)
				}
			}
		}
	}
	return nil, QueryCloseError
}
func (qv *queryValue) stop() {
	qv.once.Do(func() {
		qv.ctxCancel()
	})
}

type nodeSearch struct {
	queryTable    queryTable
	searchNode    *Node
	nodeStatus    *entity.NodeStatus
	ctxCancel     context.CancelFunc
	ctx           context.Context
	tempQueryNode *queryValue
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
	queryNode := NewQueryValue(nodeSearch.queryTable, nodeSearch.searchNode.id, nodeSearch.ctx)
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
	queryNode := NewQueryValue(nodeSearch.queryTable, nodeSearch.searchNode.id, nodeSearch.ctx)
	nodeSearch.queryNode0(queryNode)
}
func (nodeSearch *nodeSearch) queryNode0(qn *queryValue) {
	node, err := qn.StartFindValue()
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
	return nodeSearch.queryTable.FindRemoteServer(target, node, distances)
}
