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

func (call *call) register(node *LocalNode, address string) (*Node, error) {
	data, _ := json.Marshal(node)
	log.Println("/discover/register", "address:", address)
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
		id, err := hex.DecodeString(response.Data.ServerName)
		if err != nil {
			return nil, err
		}
		response.Data.SetID(wrapId(id))
		return response.Data, nil
	}
	return nil, errors.New(response.Error)
}
