package discover

import (
	"math/bits"
	"testing"
)

func TestName(t *testing.T) {

	//var data = [32]byte{128}
	//var table = NewTable(nil, &Node{id: data}, nil)
	//queryTable01 := NewQueryTable(table, &Node{id: data})
	//nodes := GenerateDistanceNodes2(data, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, 16)
	//for _, n := range nodes {
	//	queryTable01.addSeenNode(wrapNode(n))
	//	var table02 = NewTable(nil, n, nil)
	//	queryTable02 := NewQueryTable(table02, &Node{id: n.ID()})
	//	nodes02 := GenerateDistanceNodes2(n.ID(), []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, 16)
	//	for _, n02 := range nodes02 {
	//		queryTable02.addSeenNode(wrapNode(n02))
	//	}
	//	queryTable01.AddTable(n.ID(), queryTable02)
	//}
	//nodeSearch := newNodeSearch(queryTable01, [32]byte{})
	//queryNodeDone := make(chan struct{})
	//nodeSearch.queryNode(queryNodeDone)

}
func TestLogDist2(t *testing.T) {
	var data01 = [32]byte{128}
	var data02 = [32]byte{136}

	t.Log(LogDist(data01, data02))

}
func TestLogDist3(t *testing.T) {

	t.Log(bits.LeadingZeros8(136))

}
