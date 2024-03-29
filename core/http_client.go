package core

import "net"

type HttpClient struct {
	context *Context
}

func NewHttpClient(context *Context) *HttpClient {
	return &HttpClient{context: context}
}
func (request *HttpClient) GetRequest(remoteAddress *net.UDPAddr, path string) (string, error) {
	client, err := request.context.GetHttpClient(remoteAddress)
	jsonString, err := client.Get(path)
	if err != nil {
		return "", err
	}
	return jsonString, err
}
func (request *HttpClient) PostRequest(remoteAddress *net.UDPAddr, path string, json string) (string, error) {
	client, err := request.context.GetHttpClient(remoteAddress)
	jsonString, err := client.PostJsonString(path, json)
	if err != nil {
		return "", err
	}
	return jsonString, err

}
