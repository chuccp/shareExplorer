package discover

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {

	//table := GenerateTable()
	//for i := 0; i < 1000; i++ {
	//	node := wrapNode(GenerateServerNode())
	//	table.addSeenNode(node)
	//	node2 := wrapNode(GenerateServerNode())
	//	table.addSeenNode(node2)
	//}
	//node := GenerateNode()
	//nodes := table.FindValue(node.ID(), 248)
	//t.Log(nodes)

	var data = [32]byte{}
	nodes := GenerateDistanceNodes(data, 1, 5)
	for _, n := range nodes {
		var id = n.id[:]
		fmt.Printf("%08b \n", id)
	}

}
