package discover

import (
	"bytes"
	"encoding/hex"
	"math/bits"
	"net"
)

type ID [32]byte

var zeroId ID

func (n ID) IsBlank() bool {
	if bytes.Compare(n[:], zeroId[:]) == 0 {
		return true
	}
	return false
}

func wrapId(id []byte) ID {
	var d ID
	if len(id) < 32 {
		copy(d[32-len(id):], id[:])
	} else if len(id) > 32 {
		copy(d[:], id[32-len(id):])
	} else {
		copy(d[:], id[:])
	}
	return d
}
func wrapIdFName(id string) (ID, error) {
	b, err := hex.DecodeString(id)
	if err != nil {
		var id ID
		return id, err
	}
	return wrapId(b), nil
}

type Node struct {
	id          ID
	ServerName  string `json:"serverName"`
	IsServer    string `json:"isServer"`
	IsNatClient string `json:"isNatClient"`
	IsNatServer string `json:"isNatServer"`
	addr        *net.UDPAddr
}

func (n *Node) IP() net.IP {
	return n.addr.IP
}

func (n *Node) ID() ID {
	return n.id
}
func (n *Node) SetID(id ID) {
	n.id = id
}

func NewNursery(addr *net.UDPAddr) *Node {
	return &Node{addr: addr, IsNatServer: "true"}
}

type LocalNode struct {
	*Node
}

func createLocalNode(serverName string) (*LocalNode, error) {
	id, err := hex.DecodeString(serverName)
	if err != nil {
		return nil, err
	}
	return &LocalNode{&Node{ServerName: serverName, id: wrapId(id)}}, nil
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
