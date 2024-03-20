package discover

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/chuccp/shareExplorer/core"
	"math/bits"
	"net"
	"strings"
	"time"
)

type ID [32]byte

func StringToId(server string) (ID, error) {
	id, err := hex.DecodeString(server)
	return ID(id), err
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
	id              ID
	isServer        bool
	isClient        bool
	isNatServer     bool
	addr            *net.UDPAddr
	addTime         time.Time
	lastUpdateTime  time.Time
	lastRefreshTime time.Time
	errorNum        int
	liveNessChecks  int
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
func (n *Node) HasId() bool {
	return !n.id.IsBlank()
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

func NewLocalNode(id ID, config *core.ServerConfig) *Node {
	return &Node{id: id, isServer: config.IsServer(), isClient: config.IsClient(), isNatServer: config.IsNatServer()}
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
