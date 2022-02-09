package templatePlus

import (
	"html/template"
	"net/http"
	"shareExplorer/file"
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
func (tp *TemplatePlus) Parse(templatePath string) error {
	file, err := file.NewFile(templatePath)
	if err != nil {
		return err
	}
	files, err := file.ListAllFile()
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			return tp.Parse(f.Abs())
		} else {
			_, err2 := tp.template.ParseFiles(f.Abs())
			if err2 != nil {
				return err2
			}
		}
	}
	return nil
}

type rawHandle struct {
	Handle Handle
	path   string
}

func (rh *rawHandle) handleFunc(w http.ResponseWriter, r *http.Request) {

}

func Parse(templatePath string) (*TemplatePlus, error) {
	tp := new(TemplatePlus)
	return tp, tp.Parse(templatePath)
}
