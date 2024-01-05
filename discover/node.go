package discover

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/chuccp/shareExplorer/core"
	"math/bits"
	"net"
	"strings"
)

type ID [32]byte

func StringToId(server string) ID {
	id, _ := hex.DecodeString(server)
	return ID(id)
}

var zeroId ID

func (n ID) IsBlank() bool {
	if bytes.Compare(n[:], zeroId[:]) == 0 {
		return true
	}
	return false
}
func (n ID) String() string {
	return fmt.Sprintf("%x", n[:])
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
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	var node Node
	node.id = wrapId(b)
	node.isClient = strings.Contains(register.IsClient, "true")
	node.isNatServer = strings.Contains(register.IsNatServer, "true")
	node.isServer = strings.Contains(register.IsServer, "true")
	node.addr = addr
	return &node, nil
}

type Node struct {
	id          ID
	isServer    bool
	isClient    bool
	isNatServer bool
	addr        *net.UDPAddr
}

func (n *Node) IP() net.IP {
	return n.addr.IP
}

func (n *Node) IsServer() bool {
	return n.isServer
}
func (n *Node) IsClient() bool {
	return n.isClient
}
func (n *Node) IsNatServer() bool {
	return n.isNatServer
}
func (n *Node) ServerName() string {
	return hex.EncodeToString(n.id[:])
}

func (n *Node) ID() ID {
	return n.id
}
func (n *Node) SetID(id ID) {
	n.id = id
}

func NewNursery(addr *net.UDPAddr) *Node {
	return &Node{addr: addr}
}

func createLocalNode(serverName string, config *core.ServerConfig) (*Node, error) {
	id, err := hex.DecodeString(serverName)
	if err != nil {
		return nil, err
	}
	return &Node{id: wrapId(id), isServer: config.IsServer(), isClient: config.IsClient(), isNatServer: config.IsNatServer()}, nil
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
func DistCmp(target, a, b ID) int {
	for i := range target {
		da := a[i] ^ target[i]
		db := b[i] ^ target[i]
		if da > db {
			return 1
		} else if da < db {
			return -1
		}
	}
	return 0
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
