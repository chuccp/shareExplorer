package web

import "encoding/json"

type Data interface {
	string | any | *PageAble
}

type Response[T Data] struct {
	Code  int    `json:"code"`
	Data  T      `json:"data"`
	Error string `json:"error"`
}

func (t *Response[T]) IsOk() bool {
	return t.Code == 200
}
func ResponseOK[T Data](msg T) *Response[T] {
	return &Response[T]{Code: 200, Data: msg}
}
func ResponsePage(total int64, list any) *Response[*PageAble] {
	return &Response[*PageAble]{Code: 200, Data: &PageAble{Total: total, List: list}}
}
func ResponseData[T Data](data T) *Response[T] {
	return &Response[T]{Code: 200, Data: data}
}
func ResponseError(msg string) *Response[string] {
	return &Response[string]{Code: 500, Error: msg}
}
func ResponseMsg(code int, msg string) *Response[string] {
	return &Response[string]{Code: code, Error: msg}
}
func JsonToResponse[T Data](jsonString string) (*Response[T], error) {
	var response Response[T]
	err := json.Unmarshal([]byte(jsonString), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil

}
