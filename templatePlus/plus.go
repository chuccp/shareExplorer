package templatePlus

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"shareExplorer/file"
)

type plusFunc func(*Context)
type Handle func(http.ResponseWriter, *http.Request)

type MapFunc func(*Context, template.FuncMap)

type rawHandle struct {
	Handle       Handle
	path         string
	plusFunc     plusFunc
	templatePlus *TemplatePlus
}

func (rh *rawHandle) handleFunc(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r, rh.templatePlus)
	rh.plusFunc(ctx)
}

type Template struct {
	template     *template.Template
	relativePath string
	templatePath string
	text         string
	funcMap template.FuncMap
}

func (t *Template) SetFuncMap(funcMap template.FuncMap) {
	t.funcMap = funcMap
}

func (t *Template) Clone() (*Template, error) {
	tm, err := t.template.Clone()
	if err != nil {
		return nil, err
	}
	return &Template{template: tm, relativePath: t.relativePath, templatePath: t.templatePath}, nil
}

func (t *Template) Parse(file *file.File) (*Template, error) {

	relativePath, err1 := filepath.Rel(t.templatePath, file.Abs())
	if err1 != nil {
		return nil, err1
	}
	t.relativePath = relativePath
	data, err := os.ReadFile(file.Abs())
	if err != nil {
		return nil, err
	}
	t.text = string(data)
	if t.template == nil {
		t.template, err = template.New(t.relativePath).Parse(t.text)
		if err != nil {
			return nil, err
		}
	} else {
		t.template, err = t.template.New(t.relativePath).Parse(t.text)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *Template) Execute(w http.ResponseWriter, data interface{}) {
	t.template.Execute(w, data)
}

type TemplatePlus struct {
	templatePath string
	template     *Template
	templateMap  map[string]*Template
	contextMap   map[string]*string
	handleMap    map[string]*rawHandle
	mapFunc      MapFunc
	debug        bool
}

func (temp *TemplatePlus) Handle(path string, f plusFunc) (string, Handle) {
	rw := new(rawHandle)
	rw.plusFunc = f
	rw.templatePlus = temp
	temp.handleMap[path] = rw
	return path, rw.handleFunc
}
func (temp *TemplatePlus) GetTemplate(relativePath string,ctx *Context) (*Template, error) {
	if temp.debug {
		return temp.debugTemplate(relativePath,ctx)
	}
	tp, err := temp.getTemplate(relativePath)
	if err == nil {
		tem, _ := tp.template.New(relativePath).Parse(tp.text)
		return &Template{template: tem, relativePath: relativePath, templatePath: temp.templatePath, text: tp.text}, err
	}
	return tp, err
}
func (temp *TemplatePlus) debugTemplate(relativePath string,ctx *Context) (*Template, error) {




	data, err := os.ReadFile(filepath.Join(temp.templatePath, relativePath))
	if err == nil {
		abc:=func() (string,error){return "uuuuu",nil}
		var mm = &template.FuncMap{"abc": abc}
		temp.mapFunc(ctx,*mm)
		log.Println(mm,relativePath)
		tem, err4 := template.New(relativePath).Funcs(*mm).Parse(string(data))
		if err4 != nil {
			return nil, err4
		}


		for relPath, _ := range temp.templateMap {
			if relPath == relativePath {
				continue
			}
			data2, err2 := os.ReadFile(filepath.Join(temp.templatePath, relPath))
			if err2 == nil {
				tem.New(relPath).Parse(string(data2))
			}
		}






		return &Template{template: tem, relativePath: relativePath, templatePath: temp.templatePath, text: string(data)}, nil
	}
	return nil, err

}

func (temp *TemplatePlus) getTemplate(relativePath string) (*Template, error) {
	v := temp.templateMap[relativePath]
	if v == nil {
		tp, err := temp.ParseFile(filepath.Join(temp.templatePath, relativePath))
		if err != nil {
			return nil, err
		}
		return tp.Clone()
	}
	return temp.templateMap[relativePath].Clone()
}
func (temp *TemplatePlus) ParseFile(filePath string) (*Template, error) {
	f, err8 := file.NewFile(filePath)
	if err8 != nil {
		return nil, err8
	}
	tmp, err4 := temp.template.Parse(f)
	if err4 != nil {
		return nil, err4
	} else {
		temp.addTemplate(tmp.relativePath, tmp)
	}
	return tmp, nil
}
func (temp *TemplatePlus) addTemplate(relativePath string, tmp *Template) {
	temp.templateMap[relativePath] = tmp
}

func (temp *TemplatePlus) Debug(debug bool) {
	temp.debug = debug
}

func (temp *TemplatePlus) Parse(templatePath string) (*TemplatePlus, error) {
	file, err := file.NewFile(templatePath)
	if err != nil {
		return nil, err
	}
	temp.templatePath = templatePath
	files, err := file.ListAllFile()
	if err != nil {
		return nil, err
	}
	if err == nil {
		for _, f := range files {
			if f.IsDir() {
				return temp.Parse(f.Abs())
			} else {
				tmp, err4 := temp.template.Parse(f)
				if err4 != nil {
					return nil, err4
				} else {
					temp.addTemplate(tmp.relativePath, tmp)
				}
			}
		}
	}
	return temp, nil
}

func (temp *TemplatePlus) Funcs(mapFunc MapFunc) {
	temp.mapFunc = mapFunc
}

func Parse(templatePath string) (*TemplatePlus, error) {
	templatePath, _ = filepath.Abs(templatePath)
	tp := &TemplatePlus{templateMap: make(map[string]*Template), handleMap: make(map[string]*rawHandle), debug: false}
	tp.template = &Template{templatePath: templatePath}
	tp.Parse(templatePath)
	return tp, nil
}
