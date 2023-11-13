package web

import (
	"net/http"
	"net/url"
)

type ReverseResponseWriter struct {
}

func (responseWriter *ReverseResponseWriter) Header() http.Header {

	return nil
}
func (responseWriter *ReverseResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (responseWriter *ReverseResponseWriter) WriteHeader(statusCode int) {

}
func NewReverseResponseWriter() *ReverseResponseWriter {
	return &ReverseResponseWriter{}
}

type ReverseRequest struct {
	*http.Request
}

func createReverseRequest(path string) (*ReverseRequest, error) {
	var req http.Request
	req.Header = make(http.Header)
	var err error
	req.URL, err = url.Parse(path)
	if err != nil {
		return nil, err
	}
	req.Method = "GET"
	return &ReverseRequest{Request: &req}, nil
}

type ReverseResponse struct {
}

type ReverseClient struct {
	*ReverseRequest
	Response *ReverseResponseWriter
}

func (rc ReverseClient) GetReverseResponse() *ReverseResponse {

	return &ReverseResponse{}
}
func CreateReverseClient(path string) (*ReverseClient, error) {
	reverseRequest, err := createReverseRequest(path)
	if err != nil {
		return nil, err
	}
	return &ReverseClient{ReverseRequest: reverseRequest, Response: NewReverseResponseWriter()}, nil
}
