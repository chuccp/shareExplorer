package traversal

import (
	"errors"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/web"
)

type Request struct {
	context *core.Context
}

func NewRequest(context *core.Context) *Request {
	return &Request{context: context}
}
func (request *Request) RequestString(remoteAddress string, path string) (string, error) {
	client, err := request.context.GetHttpClient(remoteAddress)
	jsonString, err := client.Get(path)
	if err != nil {
		return "", err
	}
	response, err := web.JsonToResponse[string](jsonString)
	if err != nil {
		return "", err
	}
	if response.IsOk() {
		return response.Data, nil
	}
	return "", errors.New(response.Data)
}
