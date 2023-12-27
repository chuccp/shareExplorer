package discover

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
	QueryNode struct {
		FormId string `json:"formId"`
		ToId   string `json:"toId"`
		Step   int    `json:"step"`
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
		IsServer    string `json:"isServer"`
		IsClient    string `json:"isClient"`
		IsNatServer string `json:"isNatServer"`
		Address     string `json:"address"`
	}
)

func wrapResponseNode(n *Node) *ResponseNode {
	return &ResponseNode{Id: n.serverName, IsServer: n.isServer, IsNatServer: n.isNatServer, IsClient: n.isClient, Address: n.addr.String()}
}
func wrapExNode(n *node) *ExNode {
	return &ExNode{Id: n.serverName, IsServer: n.isServer, IsNatServer: n.isNatServer, IsClient: n.isClient, Address: n.addr.String()}
}

func wrapExNodes(ns []*node) []*ExNode {
	if ns == nil {
		return make([]*ExNode, 0)
	}
	var responseNodes = make([]*ExNode, len(ns))
	for i, n := range ns {
		responseNodes[i] = wrapExNode(n)
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
