package discover

import (
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
