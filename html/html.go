package html

import (
	"html/template"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	tmpl,err:=template.ParseFiles("template/index.html","template/header.html","template/end.html","template/nav.html")
	if err==nil{
		tmpl.Execute(w,"")
	}else{
		w.Write([]byte(err.Error()))
	}
}

func local(w http.ResponseWriter, r *http.Request,value map[string]interface{}) {
	tmpl,err:=template.New("local.html").Funcs(template.FuncMap{"queryUrl": func()string {
		return r.RequestURI
	}}).ParseFiles("template/local.html","template/header.html","template/end.html","template/nav.html")
	if err==nil{
		tmpl.Execute(w,"")
	}else{
		w.Write([]byte(err.Error()))
	}
}
func Html() {

	http.HandleFunc("/", index)
	http.HandleFunc("/local.html", Handle(local))
	http.Handle("/js/",  http.StripPrefix("/js/",http.FileServer(http.Dir("static/js/"))))
	http.Handle("/css/",  http.StripPrefix("/css/",http.FileServer(http.Dir("static/css/"))))
	http.Handle("/fonts/",  http.StripPrefix("/fonts/",http.FileServer(http.Dir("static/fonts/"))))
	http.Handle("/images/",  http.StripPrefix("/images/",http.FileServer(http.Dir("static/images/"))))
	http.ListenAndServe(":6363", nil)

}
