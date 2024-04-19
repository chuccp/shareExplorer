package discover

import (
	"encoding/json"
	"errors"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/web"
	"go.uber.org/zap"
	"net"
	"strconv"
)

type call struct {
	httpClient *core.HttpClient
	localNode  *Node
	context    *core.Context
}

func newCall(localNode *Node, httpClient *core.HttpClient, context *core.Context) *call {
	return &call{localNode: localNode, httpClient: httpClient, context: context}
}

func (call *call) register(remoteAddress *net.UDPAddr) (*Node, error) {
	var register = &Register{
		FormId:      call.localNode.ServerName(),
		IsServer:    strconv.FormatBool(call.localNode.isServer),
		IsNatServer: strconv.FormatBool(call.localNode.isNatServer),
		IsClient:    strconv.FormatBool(call.localNode.isClient),
	}
	data, _ := json.Marshal(register)
	call.context.GetLog().Debug("register", zap.String("send", string(data)))
	value, err := call.httpClient.PostRequest(remoteAddress, "/discover/register", string(data))
	if err != nil {
		return nil, err
	}
	call.context.GetLog().Debug("register", zap.String("receive", value))
	response, err := web.JsonToResponse[*ResponseNode](value)
	if err != nil {
		return nil, err
	}
	if response.IsOk() {
		response.Data.Address = remoteAddress.String()
		toNode, err := wrapResponseNodeToNode(response.Data)
		if err != nil {
			return nil, err
		}
		return toNode, nil
	}
	return nil, errors.New(response.Error)
}

func lookupDistances0(target, dest ID) (dists []uint) {
	td := LogDist(target, dest)
	dists = append(dists, uint(td))
	for i := 1; len(dists) < lookupRequestLimit; i++ {
		if td+i <= 256 {
			dists = append(dists, uint(td+i))
		}
		if td-i > 0 {
			dists = append(dists, uint(td-i))
		}
	}
	return dists
}

func (call *call) findNode(target ID, queryNode *Node) ([]*Node, error) {
	distances := lookupDistances0(target, queryNode.ID())
	var findNode = &FindNode{FormId: call.localNode.ServerName(), ToId: queryNode.ServerName(), Distances: distances}
	data, _ := json.Marshal(findNode)
	value, err := call.httpClient.PostRequest(queryNode.addr, "/discover/findNode", string(data))
	if err != nil {
		return nil, err
	}
	call.context.GetLog().Debug("findNode", zap.Any("value", value))
	response, err := web.JsonToResponse[[]*ResponseNode](value)
	if err != nil {
		return nil, err
	}
	if response.IsOk() {
		call.context.GetLog().Debug("findNode", zap.Any("nodes", response.Data))
		return wrapResponseNodeToNodes(response.Data), nil
	}
	return nil, errors.New(response.Error)
}
func (call *call) findServer(target ID, distances int, address *net.UDPAddr) (*Node, []*Node, error) {
	var findServer = &FindServer{FormId: call.localNode.ServerName(), Target: target.String(), Distances: distances}
	data, _ := json.Marshal(findServer)
	value, err := call.httpClient.PostRequest(address, "/discover/findServer", string(data))
	if err != nil {
		return nil, nil, err
	}
	response, err := web.JsonToResponse[*FindServerResponse](value)
	if err != nil {
		return nil, nil, err
	}
	if response.IsOk() {
		node, err := wrapResponseNodeToNode(response.Data.Server)
		if err != nil {
			return nil, nil, err
		}
		return node, wrapResponseNodeToNodes(response.Data.Nodes), nil
	}
	return nil, nil, errors.New(response.Error)
}

func (call *call) ping(address *net.UDPAddr) error {
	var queryNode = &Ping{FormId: call.localNode.ServerName()}
	data, _ := json.Marshal(queryNode)
	_, err := call.httpClient.PostRequest(address, "/discover/nodeStatus", string(data))
	if err != nil {
		return err
	}
	return nil
}
