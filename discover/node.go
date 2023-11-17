package discover

import (
	"math/bits"
	"net"
)

type ID [32]byte

type Node struct {
	id          ID
	ServerName  string `json:"serverName"`
	IsServer    string `json:"isServer"`
	IsNatClient string `json:"isNatClient"`
	IsNatServer string `json:"isNatServer"`
	addr        *net.UDPAddr
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
