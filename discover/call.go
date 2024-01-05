package discover

import (
	"encoding/json"
	"errors"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/web"
	"log"
	"net"
)

type call struct {
	httpClient *core.HttpClient
}

func (call *call) register(localNode *Node, address *net.UDPAddr) (*Node, error) {
	log.Println("register==", "serverName:", localNode.ServerName(), " IsServer:", localNode.isServer, " IsNatServer:", localNode.isNatServer, " IsClient:", localNode.isClient)
	data, _ := json.Marshal(NodeToRegister(localNode))
	value, err := call.httpClient.PostRequest(address, "/discover/register", string(data))
	if err != nil {
		return nil, err
	}
	response, err := web.JsonToResponse[*ResponseNode](value)
	if err != nil {
		return nil, err
	}
	if response.IsOk() {
		response.Data.Address = address.String()
		toNode, err := wrapResponseNodeToNode(response.Data)
		if err != nil {
			return nil, err
		}
		return toNode, nil
	}
	return nil, errors.New(response.Error)
}
func (call *call) findNode(node *Node, toNode *Node, address *net.UDPAddr, distances []uint) ([]*Node, error) {
	var queryNode = &FindNode{FormId: node.ServerName(), ToId: toNode.ServerName(), Distances: distances}
	data, _ := json.Marshal(queryNode)
	value, err := call.httpClient.PostRequest(address, "/discover/queryNode", string(data))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	response, err := web.JsonToResponse[[]*Node](value)
	if err != nil {
		return nil, err
	}
	if response.IsOk() {
		return response.Data, nil
	}
	return nil, errors.New(response.Error)
}
func (call *call) findValue(target string, distances int, address *net.UDPAddr) ([]*Node, error) {
	var queryNode = &FindValue{Target: target, Distances: distances}
	data, _ := json.Marshal(queryNode)
	value, err := call.httpClient.PostRequest(address, "/discover/findValue", string(data))
	if err != nil {
		return nil, err
	}
	response, err := web.JsonToResponse[[]*Node](value)
	if err != nil {
		return nil, err
	}
	if response.IsOk() {
		return response.Data, nil
	}
	return nil, errors.New(response.Error)
}

func (call *call) ping(node *Node, address *net.UDPAddr) error {
	var queryNode = &NodeStatus{FormId: node.ServerName()}
	data, _ := json.Marshal(queryNode)
	_, err := call.httpClient.PostRequest(address, "/discover/nodeStatus", string(data))
	if err != nil {
		return err
	}
	return nil
}
