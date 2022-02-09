package templatePlus

import (
	"html/template"
	"net/http"
)

type Context struct {
	w http.ResponseWriter
	r *http.Request
	template *template.Template
}

func (ctx *Context)MajorPath(templateFile string)  {
	tmpl:=ctx.template.New(templateFile)
	tmpl.Execute(ctx.w, "")
}
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := new(Context)
	ctx.w = w
	ctx.r = r
	return ctx
}
