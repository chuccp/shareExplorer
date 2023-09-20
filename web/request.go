package web

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"strings"
)

type Request struct {
	context *gin.Context
}

func NewRequest(context *gin.Context) *Request {
	return &Request{context: context}
}

func (r *Request) FormValue(key string) string {
	return r.context.Request.FormValue(key)
}
func (r *Request) FormIntValue(key string) int {
	v := r.FormValue(key)
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return i
}
func (r *Request) GetRemoteAddress() string {
	address := r.context.Request.RemoteAddr
	index := strings.Index(address, "_")
	if index > 0 {
		return address[:index]
	}
	return address
}

func (r *Request) GetPage() *Page {
	var page Page
	page.PageNo = r.FormIntValue("pageNo")
	page.PageSize = r.FormIntValue("pageSize")
	return &page
}

func (r *Request) BodyJson(v any) error {
	log.Println(r.context.Request)
	body, err := io.ReadAll(r.context.Request.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}
	return nil
}

func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	return r.context.FormFile(name)
}
