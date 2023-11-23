package discover

import (
	"bytes"
	"encoding/hex"
	"math/bits"
	"net"
	"strings"
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
func wrapNodeFRegister(register *Register, address string) (*Node, error) {
	b, err := hex.DecodeString(register.FormId)
	if err != nil {
		return nil, err
	}
	var node Node
	node.id = wrapId(b)
	node.serverName = register.FormId
	node.isNatClient = register.IsNatClient
	node.isNatServer = register.IsNatServer
	node.isServer = register.IsServer
	return &node, nil
}

type Node struct {
	id          ID
	serverName  string
	isServer    string
	isNatClient string
	isNatServer string
	addr        *net.UDPAddr
}

func (n *Node) IP() net.IP {
	return n.addr.IP
}

func (n *Node) IsServer() bool {
	return strings.Contains(n.isServer, "true")
}
func (n *Node) IsNatClient() bool {
	return strings.Contains(n.isNatClient, "true")
}
func (n *Node) IsNatServer() bool {
	return strings.Contains(n.isNatServer, "true")
}

func (n *Node) ID() ID {
	return n.id
}
func (n *Node) SetID(id ID) {
	n.id = id
}

func NewNursery(addr *net.UDPAddr) *Node {
	return &Node{addr: addr, isNatServer: "true", isServer: "true", isNatClient: "true"}
}

func createLocalNode(serverName string) (*Node, error) {
	id, err := hex.DecodeString(serverName)
	if err != nil {
		return nil, err
	}
	return &Node{serverName: serverName, id: wrapId(id)}, nil
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
func unwrapNodes(ns []*node) []*Node {
	result := make([]*Node, len(ns))
	for i, n := range ns {
		result[i] = unwrapNode(n)
	}
	return result
}
func unwrapNode(n *node) *Node {
	return &n.Node
}
