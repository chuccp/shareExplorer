package templatePlus

import (
	"log"
	"net/http"
)

type Context struct {
	w http.ResponseWriter
	r *http.Request
	temPlus *TemplatePlus
}

func (ctx *Context) Major(templateFile string)  {
	temp,err :=ctx.temPlus.GetTemplate(templateFile)
	if err==nil{
		temp.Execute(ctx.w, "")
	}else{
		log.Println(err)
		ctx.w.Write([]byte(err.Error()))
	}
}
func NewContext(w http.ResponseWriter, r *http.Request,temp *TemplatePlus) *Context {
	ctx := new(Context)
	ctx.w = w
	ctx.r = r
	ctx.temPlus = temp
	return ctx
}
