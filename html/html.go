package html

import (
	"html/template"
	"net/http"
	"shareExplorer/templatePlus"
)

func index(w http.ResponseWriter, r *http.Request) {
//	tmpl, err := template.ParseFiles("template/index.html", "template/header.html", "template/end.html", "template/nav.html")

	tmpl,err := template.New("local.html").ParseFiles("tlocal.html")

	if err == nil {
		tmpl.Execute(w, "")
	} else {
		w.Write([]byte(err.Error()))
	}
}

func local(ctx *templatePlus.Context) {

	ctx.Major("local.html")
}
func Html() {

	template, _ := templatePlus.Parse("template")
	http.HandleFunc("/", index)

	http.HandleFunc(template.Handle("/local.html", local))

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("static/js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css/"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("static/fonts/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("static/images/"))))
	http.ListenAndServe(":6363", nil)

}
