package discover

type (
	Register struct {
		FormId      string `json:"formId"`
		IsServer    string `json:"isServer"`
		IsNatClient string `json:"isNatClient"`
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
	}
	ResponseNode struct {
		Id          string `json:"id"`
		IsServer    string `json:"isServer"`
		IsNatClient string `json:"isNatClient"`
		IsNatServer string `json:"isNatServer"`
		Address     string `json:"address"`
	}
)

func wrapResponseNode(n *Node) *ResponseNode {
	return &ResponseNode{Id: n.serverName, IsServer: n.isServer, IsNatServer: n.isNatServer, IsNatClient: n.isNatClient, Address: n.addr.String()}
}
func wrapResponseNodes(ns []*Node) []*ResponseNode {
	var responseNodes = make([]*ResponseNode, len(ns))
	for i, n := range ns {
		responseNodes[i] = wrapResponseNode(n)
	}
	return responseNodes
}
