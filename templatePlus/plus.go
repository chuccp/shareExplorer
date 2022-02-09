package templatePlus

import (
	"html/template"
	"net/http"
)

type plusFunc func(*Context)
type Handle func(http.ResponseWriter, *http.Request)

type TemplatePlus struct {
	plusFunc  plusFunc
	handleMap map[string]*rawHandle
	template  *template.Template
}

func (tp *TemplatePlus) Handle(path string, f plusFunc) (string, Handle) {
	rw := new(rawHandle)
	tp.handleMap[path] = rw
	return path, rw.handleFunc
}

type rawHandle struct {
	Handle Handle
	path   string
}

func (rh *rawHandle) handleFunc(w http.ResponseWriter, r *http.Request) {

}

func Parse(templatePath string) (*TemplatePlus, error) {
	tp := new(TemplatePlus)
	var err error
	tp.template, err = template.ParseFiles("")
	return tp, err
}
