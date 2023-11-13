package traversal

import (
	"github.com/chuccp/shareExplorer/core"
	"log"
	"time"
)

type ClientManager struct {
	context              *core.Context
	traversalHostMap     map[string]*Client
	tempTraversalHostMap map[string]*Client
}

func NewClientManager(context *core.Context) *ClientManager {
	return &ClientManager{context: context, traversalHostMap: make(map[string]*Client), tempTraversalHostMap: make(map[string]*Client)}
}

func (clientManager *ClientManager) Run() {
	addresses, err := clientManager.context.GetDB().GetAddressModel().QueryAddresses()
	if err != nil {
		log.Println(err)
	} else {
		for _, address := range addresses {
			client := NewClient(address.Address, clientManager.context)
			clientManager.tempTraversalHostMap[address.Address] = client
		}
		clientManager.register()
		clientManager.live()
	}
}

func (clientManager *ClientManager) live() {
	for {
		clientManager.readNode()
		time.Sleep(time.Second * 30)
		clientManager.register()
	}
}
func (clientManager *ClientManager) register() {
	for address, client := range clientManager.tempTraversalHostMap {
		err := client.register()
		if err == nil {
			clientManager.traversalHostMap[address] = client
		}
	}
}
func (clientManager *ClientManager) readNode() {
	for _, client := range clientManager.traversalHostMap {
		client.readNode()
	}
}

func (clientManager *ClientManager) FindRemoteHost(serverName string) {
	for _, client := range clientManager.traversalHostMap {
		client.FindRemoteHost(serverName)
	}
}

type Client struct {
	context       *core.Context
	remoteAddress string
	failureNum    int
}

func NewClient(remoteAddress string, context *core.Context) *Client {
	return &Client{remoteAddress: remoteAddress, context: context, failureNum: 0}
}
func (cl *Client) register() error {

	return nil
}
func (cl *Client) readNode() error {

	return nil
}
func (cl *Client) FindRemoteHost(serverName string) error {

	return nil
}
