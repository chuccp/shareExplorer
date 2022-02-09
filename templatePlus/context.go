package templatePlus

import "net/http"

type Context struct {
	w http.ResponseWriter
	r *http.Request
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := new(Context)
	ctx.w = w
	ctx.r = r
	return ctx
}
