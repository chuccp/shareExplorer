package discover

import (
	"encoding/hex"
	"net"
	"strconv"
	"strings"
)

type (
	Register struct {
		FormId      string `json:"formId"`
		IsServer    string `json:"isServer"`
		IsClient    string `json:"isClient"`
		IsNatServer string `json:"isNatServer"`
	}

	FindNode struct {
		FormId    string `json:"formId"`
		ToId      string `json:"toId"`
		Distances []uint `json:"distances"`
	}
	FindValue struct {
		Target    string `json:"target"`
		Distances int    `json:"distances"`
	}

	NodeStatus struct {
		FormId string `json:"id"`
	}

	ResponseNode struct {
		Id          string `json:"id"`
		IsServer    string `json:"isServer"`
		IsClient    string `json:"isClient"`
		IsNatServer string `json:"isNatServer"`
		Address     string `json:"address"`
	}

	ExNode struct {
		Id          string `json:"id"`
		IsServer    bool   `json:"isServer"`
		IsClient    bool   `json:"isClient"`
		IsNatServer bool   `json:"isNatServer"`
		Address     string `json:"address"`
	}
)

func NodeToRegister(n *Node) *Register {
	var register = &Register{FormId: n.ServerName(), IsServer: strconv.FormatBool(n.isServer), IsNatServer: strconv.FormatBool(n.isNatServer), IsClient: strconv.FormatBool(n.isClient)}
	return register
}
func wrapResponseNode(n *Node) *ResponseNode {
	address := ""
	if n.addr != nil {
		address = n.addr.String()
	}
	return &ResponseNode{Id: hex.EncodeToString(n.id[:]), IsServer: strconv.FormatBool(n.isServer), IsNatServer: strconv.FormatBool(n.isNatServer), IsClient: strconv.FormatBool(n.isClient), Address: address}
}

func wrapResponseNodeToNode(n *ResponseNode) (*Node, error) {
	id, err := hex.DecodeString(n.Id)
	if err != nil {
		return nil, err
	}
	addr, err := net.ResolveUDPAddr("udp", n.Address)
	if err != nil {
		return nil, err
	}
	return &Node{addr: addr, id: ID(id), isClient: strings.Contains(n.IsClient, "true"), isServer: strings.Contains(n.IsServer, "true"), isNatServer: strings.Contains(n.IsNatServer, "true")}, nil
}

func wrapExNode(n *Node) *ExNode {
	return &ExNode{Id: hex.EncodeToString(n.id[:]), IsServer: n.isServer, IsNatServer: n.isNatServer, IsClient: n.isClient, Address: n.addr.String()}
}

func wrapExNodes(ns []*node) []*ExNode {
	if ns == nil {
		return make([]*ExNode, 0)
	}
	var responseNodes = make([]*ExNode, len(ns))
	for i, n := range ns {
		responseNodes[i] = wrapExNode(&n.Node)
	}
	return responseNodes
}

func wrapResponseNodes(ns []*Node) []*ResponseNode {
	var responseNodes = make([]*ResponseNode, len(ns))
	for i, n := range ns {
		responseNodes[i] = wrapResponseNode(n)
	}
	return responseNodes
}
