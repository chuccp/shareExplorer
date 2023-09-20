package web

type Response struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

func ResponseOK(msg string) *Response {
	return &Response{Code: 200, Data: msg}
}
func ResponseData(data any) *Response {
	return &Response{Code: 200, Data: data}
}
