package web

import (
	"encoding/json"
	"github.com/chuccp/shareExplorer/util"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type Request struct {
	context *gin.Context
	jwt     *util.Jwt
}

func NewRequest(context *gin.Context, jwt *util.Jwt) *Request {
	return &Request{context: context, jwt: jwt}
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
func (r *Request) GetTokenUsername() string {
	token := r.context.GetHeader("Token")
	if len(token) == 0 {
		token = r.context.Request.FormValue("Token")
	}
	log.Println("token:", token)
	if len(token) > 0 {
		sub, err := r.jwt.ParseWithSub(token)
		log.Println(sub, err)
		if err != nil {
			return ""
		} else {
			return sub
		}
	}
	return ""
}
func (r *Request) SignedUsername(username string) (string, error) {
	return r.jwt.SignedSub(username)
}

func (r *Request) GetPage() *Page {
	var page Page
	page.PageNo = r.FormIntValue("pageNo")
	page.PageSize = r.FormIntValue("pageSize")
	return &page
}
func (r *Request) GetRawRequest() *http.Request {
	return r.context.Request
}

func (r *Request) BodyJson(v any) error {
	log.Println(r.context.Request)
	body, err := io.ReadAll(r.context.Request.Body)
	if err != nil {
		return err
	}
	println(string(body))
	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}
	return nil
}

func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	return r.context.FormFile(name)
}
