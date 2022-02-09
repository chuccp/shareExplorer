package templatePlus

import "net/http"

type plusFunc func(*Context)
type Handle func(http.ResponseWriter, *http.Request)

type TemplatePlus struct {
	plusFunc plusFunc
	handleMap map[string]*rawHandle
}

func (tp *TemplatePlus) Handle(path string,f plusFunc) (string,Handle) {
	rw:=new(rawHandle)
	tp.handleMap[path] = rw
	return path,rw.handleFunc
}

type rawHandle struct {
	Handle Handle
	path string
}
func (rh *rawHandle) handleFunc(w http.ResponseWriter, r *http.Request) {


}

func New(templatePath string) *TemplatePlus {
	tp := new(TemplatePlus)
	return tp
}
