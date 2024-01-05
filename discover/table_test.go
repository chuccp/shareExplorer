package discover

import (
	"crypto/rand"
	"fmt"
	rand0 "math/rand"
	"net"
	"testing"
)

func GenerateNode() *Node {
	var data [32]byte
	rand.Read(data[:])
	return &Node{id: data, isServer: true, isNatServer: true, addr: GenerateUDPAddr()}
}
func GenerateServerNode() *Node {
	var data [32]byte
	rand.Read(data[:])
	return &Node{id: data, isServer: true, isNatServer: true, addr: GenerateUDPAddr()}
}
func GenerateNodes(num int) []*Node {
	var nodes = make([]*Node, num)
	for i := 0; i < num; i++ {
		var data [32]byte
		rand.Read(data[:])
		nodes[i] = GenerateNode()
	}
	return nodes
}

func TestTable_AddSeenNodes(t *testing.T) {
	table := GenerateTable()
	for i := 0; i < 1000; i++ {
		node := wrapNode(GenerateServerNode())
		table.addSeenNode(node)
		node2 := wrapNode(GenerateServerNode())
		table.addSeenNode(node2)
	}
	node := GenerateNode()
	nodes := table.FindValue(node.ServerName(), 248)
	t.Log(nodes)
}

func TestTable_AddSeenNode(t *testing.T) {
	table := GenerateTable()
	node := wrapNode(GenerateServerNode())
	table.addSeenNode(node)
	node0, fa := table.queryServerNode(node.ServerName())
	t.Log(node0, fa)

}

func GenerateServerNodes(num int) []*Node {
	var nodes = make([]*Node, num)
	for i := 0; i < num; i++ {
		nodes[i] = GenerateServerNode()
	}
	return nodes
}

func GenerateTable() *Table {
	node := GenerateNode()
	return NewTable(nil, node, nil)
}
func randIPv4() string {
	var ipBytes [4]byte
	for i := 0; i < 4; i++ {
		ipBytes[i] = byte(rand0.Intn(256))
	}
	return fmt.Sprintf("%d.%d.%d.%d", ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])
}
func randPort() int {
	return 1024 + rand0.Intn(65535-1024)
}
func GenerateUDPAddr() *net.UDPAddr {
	ip := randIPv4()
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: randPort(),
	}
	return udpAddr
}
