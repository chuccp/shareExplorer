package html

import (
	"html/template"
	"net/http"
	"shareExplorer/templatePlus"
)

func index(ctx *templatePlus.Context) {
//	tmpl, err := template.ParseFiles("template/index.html", "template/header.html", "template/end.html", "template/nav.html")
	ctx.Major("index.html")
}

func local(ctx *templatePlus.Context) {

	ctx.Major("local.html")
}
func Html() {

	temp, _ := templatePlus.Parse("template")

	temp.Funcs(func(context *templatePlus.Context, funcMap template.FuncMap) {
		funcMap["requestUrl"] = func(iii string)(string,error) {
			return "123123"+iii,nil
		}
	})

	temp.Debug(true)
	//http.HandleFunc(temp.Handle("/", index))
	http.HandleFunc(temp.Handle("/local.html", local))

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("static/js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css/"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("static/fonts/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("static/images/"))))
	http.ListenAndServe(":6363", nil)

}
