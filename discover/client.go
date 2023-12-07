package discover

import (
	"container/list"
	"time"
)

type nodeSearch struct {
	table      *Table
	localNode  *Node
	err        error
	nodeStatus int
}

func newNodeStatus(table *Table, localNode *Node) *nodeSearch {
	return &nodeSearch{table: table, localNode: localNode}
}
func (nodeSearch *nodeSearch) run() {
	go nodeSearch.loop()
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

type NodeMaxBucket struct {
	node      *Node
	maxBucket uint
}

type findValueNode struct {
	useNode map[string]*Node
	node    *list.List
}

func (f *findValueNode) addNode(target ID, node *Node) {

}
func (f *findValueNode) getNode(maxnum int) ([]*NodeMaxBucket, bool) {

	return nil, false
}
func NewFindValueNode() *findValueNode {
	return &findValueNode{useNode: make(map[string]*Node), node: new(list.List)}
}

func (nodeSearch *nodeSearch) queryNode(done chan<- struct{}) {
	defer close(done)
	var findValueNode = NewFindValueNode()
	queryNode, _, err := nodeSearch.table.FindValue(nodeSearch.localNode.serverName, nBuckets)
	if err != nil {
		return
	}
	for _, n := range queryNode {
		findValueNode.addNode(nodeSearch.localNode.id, n)
	}
	for {
		nodes, fa := findValueNode.getNode(16)
		if fa {
			break
		}
		for _, n := range nodes {
			queryNode, local, err := nodeSearch.FindValue(nodeSearch.localNode.serverName, n.node, n.maxBucket)
			if err == nil {
				if local != nil {
					err := nodeSearch.ping(local)
					if err != nil {
						nodeSearch.err = err
					}
					nodeSearch.localNode.addr = local.addr
					break
				} else {
					for _, n := range queryNode {
						findValueNode.addNode(nodeSearch.localNode.id, n)
					}
				}
			}
		}
	}
}
func (nodeSearch *nodeSearch) ping(node *Node) error {

	return nil
}

func (nodeSearch *nodeSearch) FindValue(serverName string, node *Node, maxBucket uint) (queryNode []*Node, local *Node, err error) {

	return
}
