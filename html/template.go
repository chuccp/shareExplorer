package html

import "net/http"

type handle func( http.ResponseWriter,  *http.Request)

type plusFunc func( http.ResponseWriter, *http.Request,  map[string]interface{})

type TemplatePlus struct {
	plusHandle plusFunc
}
func (tp *TemplatePlus) Handle(w http.ResponseWriter, r *http.Request) {
	var mm = make(map[string]interface{})
	tp.plusHandle(w, r, mm)
}

func Handle(f plusFunc) handle {
	tp := new(TemplatePlus)
	tp.plusHandle = f
	return tp.Handle
}
