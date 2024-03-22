package discover

import (
	"encoding/hex"
	"github.com/chuccp/shareExplorer/util"
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
		TargetId  string `json:"targetId"`
		Distances []uint `json:"distances"`
		addr      *net.UDPAddr
	}
	FindServer struct {
		FormId    string `json:"formId"`
		Target    string `json:"target"`
		Distances int    `json:"distances"`
	}

	FindServerResponse struct {
		Server *ResponseNode   `json:"server"`
		Nodes  []*ResponseNode `json:"nodes"`
	}

	NodeStatus struct {
		Id          string `json:"id"`
		IsServer    string `json:"isServer"`
		IsClient    string `json:"isClient"`
		IsNatServer string `json:"isNatServer"`
	}
	Ping struct {
		FormId      string `json:"formId"`
		IsServer    string `json:"isServer"`
		IsClient    string `json:"isClient"`
		IsNatServer string `json:"isNatServer"`
	}

	ResponseNode struct {
		Id          string `json:"id"`
		IsServer    string `json:"isServer"`
		IsClient    string `json:"isClient"`
		IsNatServer string `json:"isNatServer"`
		Address     string `json:"address"`
	}

	ExNode struct {
		Id           string `json:"id"`
		Address      string `json:"address"`
		LastLiveTime string `json:"lastLiveTime"`
	}
)

func NodeToRegister(n *Node) *Register {
	var register = &Register{FormId: n.ServerName(), IsServer: strconv.FormatBool(n.isServer), IsNatServer: strconv.FormatBool(n.isNatServer), IsClient: strconv.FormatBool(n.isClient)}
	return register
}
func NodeToExNode(n *Node) *ExNode {
	var register = &ExNode{Id: n.ServerName(), Address: n.addr.String(), LastLiveTime: util.FormatTime(&n.lastUpdateTime)}
	return register
}
func NodeToExNodes(ns []*Node) []*ExNode {
	var nodes = make([]*ExNode, len(ns))
	for i, n := range ns {
		nodes[i] = NodeToExNode(n)
	}
	return nodes
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

func wrapResponseNodeToNodes(ns []*ResponseNode) []*Node {
	var nodes = make([]*Node, 0)
	for _, n := range ns {
		node, err := wrapResponseNodeToNode(n)
		if err == nil {
			nodes = append(nodes, node)
		}
	}
	return nodes

}

func wrapFindServerResponse(n *Node, ns []*Node) *FindServerResponse {
	return &FindServerResponse{Server: wrapResponseNode(n), Nodes: wrapResponseNodes(ns)}
}

func wrapResponseNodes(ns []*Node) []*ResponseNode {
	var responseNodes = make([]*ResponseNode, len(ns))
	for i, n := range ns {
		responseNodes[i] = wrapResponseNode(n)
	}
	return responseNodes
}
