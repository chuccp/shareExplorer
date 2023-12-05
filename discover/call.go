package discover

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/web"
	"log"
)

type call struct {
	httpClient *core.HttpClient
}

func (call *call) register(node *Node, address string) (*Node, error) {
	var register = &Register{FormId: node.serverName, IsServer: node.isServer, IsNatServer: node.isNatServer, IsNatClient: node.isNatClient}
	data, _ := json.Marshal(register)
	value, err := call.httpClient.PostRequest(address, "/discover/register", string(data))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	response, err := web.JsonToResponse[*Node](value)
	if err != nil {
		return nil, err
	}
	if response.IsOk() {
		id, err := hex.DecodeString(response.Data.serverName)
		if err != nil {
			return nil, err
		}
		response.Data.SetID(wrapId(id))
		return response.Data, nil
	}
	return nil, errors.New(response.Error)
}
func (call *call) findNode(node *Node, toNode *Node, address string, distances []uint) ([]*Node, error) {
	var queryNode = &FindNode{FormId: node.serverName, ToId: toNode.serverName, Distances: distances}
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
func (call *call) queryNode(node *Node, toNode *Node, address string) (*Node, error) {
	var queryNode = &QueryNode{FormId: node.serverName, ToId: toNode.serverName}
	data, _ := json.Marshal(queryNode)
	value, err := call.httpClient.PostRequest(address, "/discover/queryNode", string(data))
	if err != nil {
		return nil, err
	}
	response, err := web.JsonToResponse[*Node](value)
	if err != nil {
		return nil, err
	}
	if response.IsOk() {
		return response.Data, nil
	}
	return nil, errors.New(response.Error)
}
