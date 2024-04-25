package web

import (
	"bytes"
	"net/http"
)

type ProxyResponse struct {
	http.ResponseWriter
	header http.Header
	buffer *bytes.Buffer
}

func NewProxyResponse(writer http.ResponseWriter) *ProxyResponse {
	return &ProxyResponse{
		ResponseWriter: writer,
		header:         http.Header{},
		buffer:         new(bytes.Buffer),
	}
}

func (proxyResponse *ProxyResponse) Header() http.Header {
	return proxyResponse.header
}
func (proxyResponse *ProxyResponse) GetBody() string {
	return proxyResponse.buffer.String()
}
func (proxyResponse *ProxyResponse) GetBytesBody() []byte {
	return proxyResponse.buffer.Bytes()
}
func (proxyResponse *ProxyResponse) Write(data []byte) (int, error) {
	return proxyResponse.buffer.Write(data)
}

func (proxyResponse *ProxyResponse) Flush() (int, error) {
	return proxyResponse.ResponseWriter.Write(proxyResponse.buffer.Bytes())
}

func (proxyResponse *ProxyResponse) WriteHeader(statusCode int) {
	proxyResponse.ResponseWriter.WriteHeader(statusCode)
}
