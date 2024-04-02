package web

import (
	"encoding/json"
	auth "github.com/abbot/go-http-auth"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type HandlerFunc func(req *Request) (any, error)

type Request struct {
	context     *gin.Context
	authRequest *auth.AuthenticatedRequest
}

func NewRequest(context *gin.Context, authRequest *auth.AuthenticatedRequest) *Request {
	return &Request{context: context, authRequest: authRequest}
}
func (r *Request) GetAuthRequest() *auth.AuthenticatedRequest {
	return r.authRequest
}
func (r *Request) GetAuthUsername() string {
	if r.authRequest != nil {
		return r.authRequest.Username
	}
	return ""
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
func (r *Request) FormInt64Value(key string) int64 {
	v := r.FormValue(key)
	i, err := strconv.ParseInt(v, 10, 64)
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
func (r *Request) GetRawRequest() *http.Request {
	return r.context.Request
}
func (r *Request) GetResponseWriter() http.ResponseWriter {
	return r.context.Writer
}

func (r *Request) BodyJson(v any) ([]byte, error) {
	body, err := io.ReadAll(r.context.Request.Body)
	if err != nil {
		return body, err
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return body, err
	}
	return body, nil
}

func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	return r.context.FormFile(name)
}

func (r *Request) BasicAuth() (username, password string, ok bool) {
	return r.context.Request.BasicAuth()
}
func (r *Request) Header(key, value string) {
	r.context.Header(key, value)
}
func (r *Request) Status(code int) {
	r.context.Status(code)
}
func (r *Request) String(code int, format string, values ...any) {
	r.context.String(code, format, values...)
}

func ToGinHandlerFuncs(handlers []HandlerFunc) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = ToGinHandlerFunc(handler)
	}
	return handlerFunc
}
func ToGinHandlerFunc(handler HandlerFunc) gin.HandlerFunc {
	handlerFunc := func(context *gin.Context) {
		value, err := handler(NewRequest(context, nil))
		if err != nil {
			context.AbortWithStatusJSON(200, ResponseError(err.Error()))
		} else {
			if value != nil {
				switch t := value.(type) {
				case string:
					context.Writer.Write([]byte(t))
				case *File:
					context.FileAttachment(t.GetPath(), t.GetFilename())
				default:
					context.AbortWithStatusJSON(200, t)
				}
			}
		}
	}
	return handlerFunc
}
