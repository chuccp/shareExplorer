package web

import (
	"errors"
	"github.com/chuccp/shareExplorer/core"
)

type HttpClient struct {
	context *core.Context
}

func NewHttpClient(context *core.Context) *HttpClient {
	return &HttpClient{context: context}
}
func (request *HttpClient) GetRequest(remoteAddress string, path string) (string, error) {
	client, err := request.context.GetHttpClient(remoteAddress)
	jsonString, err := client.Get(path)
	if err != nil {
		return "", err
	}
	response, err := JsonToResponse[string](jsonString)
	if err != nil {
		return "", err
	}
	if response.IsOk() {
		return response.Data, nil
	}
	return "", errors.New(response.Data)
}
func (request *HttpClient) PostRequest(remoteAddress string, path string, json string) (string, error) {
	client, err := request.context.GetHttpClient(remoteAddress)
	jsonString, err := client.PostJsonString(path, json)
	if err != nil {
		return "", err
	}
	response, err := JsonToResponse[string](jsonString)
	if err != nil {
		return "", err
	}
	if response.IsOk() {
		return response.Data, nil
	}
	return "", errors.New(response.Data)
}
