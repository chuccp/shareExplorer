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
func GenerateZeroNode() *Node {
	var data = [32]byte{}
	return &Node{id: data, isServer: true, isNatServer: true, addr: GenerateUDPAddr()}
}

func GenerateDistanceNodes(id ID, nBuckets int, num int) []*Node {
	v := ReadBit(id[:], 16-nBuckets)
	nodes := GenerateNodes(num)
	for _, n := range nodes {
		//fmt.Printf("%08b \n", n.id[:])
		BitReplace(id[:], n.id[:], uint(16-nBuckets), 0)
		n.id = id
		SetBit(n.id[:], 16-nBuckets, !v)
		//fmt.Printf("%08b \n", n.id[:])
	}
	return nodes
}

func ReadBit(src []byte, index int) bool {
	startByteIndex := (index) / 8
	v := src[startByteIndex]
	var bitIndex = index % 8
	return v<<bitIndex>>bitIndex > 0
}
func SetBit(src []byte, index int, value bool) {
	startByteIndex := (index) / 8
	var bitIndex = index % 8
	var v1 byte = 0b11111111
	v1 = v1 << bitIndex >> bitIndex
	v1 = v1 >> (8 - (bitIndex + 1)) << (8 - (bitIndex + 1))
	if value {
		src[startByteIndex] = src[startByteIndex] | v1
	} else {
		src[startByteIndex] = src[startByteIndex] & (^v1)
	}
}

func BitReplace(src []byte, replaceWith []byte, start uint, bitCount uint) {

	startByteIndex := int((start) / 8)
	if bitCount == 0 {
		bitCount = uint(uint(len(src)*8) - uint(start))
	}
	endByteIndex := int((bitCount + start) / 8)

	if (bitCount+start)%8 == 0 {
		endByteIndex = endByteIndex - 1
	}

	for i := startByteIndex; i <= endByteIndex; i++ {
		if i == startByteIndex && i != endByteIndex {
			var bitIndex = start % 8
			var v1 = replaceWith[i] << bitIndex >> bitIndex
			var v2 = src[i] >> (8 - bitIndex) << (8 - bitIndex)
			src[i] = v1 | v2

		} else if i == startByteIndex && i == endByteIndex {
			var bitIndex = start % 8
			var bitIndex2 = (bitCount + start) % 8
			var v byte = 0b11111111
			v = v << bitIndex >> bitIndex
			v = v >> (8 - bitIndex2) << (8 - bitIndex2)
			var v2 = v & replaceWith[i]
			var v1 = (^v) & src[i]
			src[i] = v1 | v2
		} else if i == endByteIndex {
			var bitIndex = (bitCount + start) % 8
			var v1 = src[i] << bitIndex >> bitIndex
			var v2 = replaceWith[i] >> (8 - bitIndex) << (8 - bitIndex)
			src[i] = v1 | v2
		} else {
			src[i] = replaceWith[i]
		}
	}
}

func TestReplace(t *testing.T) {
	src := []byte{0b00000000, 0b00000000, 0b00000000}
	replaceWith := []byte{0b11111111, 0b11111111, 0b11111111}
	BitReplace(src, replaceWith, 1, 8)
	fmt.Printf("%08b ", src)
}

func TestReadBit(t *testing.T) {
	src := []byte{0b10000000, 0b00000000, 0b00000000}

	t.Log(ReadBit(src, 8))
}
func TestSetBit(t *testing.T) {
	src := []byte{0b00000000, 0b00000000, 0b00000000}
	SetBit(src, 10, true)
	fmt.Printf("%08b ", src)
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
	nodes := table.FindValue(node.ID(), 248)
	t.Log(nodes)
}

func TestTable_AddSeenNode(t *testing.T) {
	table := GenerateTable()
	node := wrapNode(GenerateServerNode())
	table.addSeenNode(node)
	node0, fa := table.queryServerNode(node.ID())
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
