package discover

import (
	"bytes"
	"math/bits"
	"net"
)

type ID [32]byte

var zeroId ID

func (n *ID) IsBlank() bool {
	if bytes.Compare(n[:], zeroId[:]) == 0 {
		return true
	}
	return false
}

type Node struct {
	id          ID
	ServerName  string `json:"serverName"`
	IsServer    string `json:"isServer"`
	IsNatClient string `json:"isNatClient"`
	IsNatServer string `json:"isNatServer"`
	addr        *net.UDPAddr
}

func (n *Node) ID() ID {
	return n.id
}

func NewNursery(addr *net.UDPAddr) *Node {
	return &Node{addr: addr, IsNatServer: "true"}
}

type LocalNode struct {
	id          ID
	ServerName  string `json:"serverName"`
	IsServer    string `json:"isServer"`
	IsNatClient string `json:"isNatClient"`
	IsNatServer string `json:"isNatServer"`
}

func newLocalNode(serverName string) *LocalNode {

	return &LocalNode{ServerName: serverName}
}

func LogDist(a, b ID) int {
	lz := 0
	for i := range a {
		x := a[i] ^ b[i]
		if x == 0 {
			lz += 8
		} else {
			lz += bits.LeadingZeros8(x)
			break
		}
	}
	return len(a)*8 - lz
}
