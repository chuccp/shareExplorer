package http

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"shareExplorer/file"
)

func index(context *gin.Context) {

	context.HTML(http.StatusOK, "index.tmpl", gin.H{})
}
func main(context *gin.Context) {

	context.HTML(http.StatusOK, "lyear_main.tmpl", gin.H{})
}
func disk(context *gin.Context) {
	files, err := file.GetRootPath()
	if err == nil {
		log.Println(len(files))
		context.HTML(http.StatusOK, "disk.tmpl", gin.H{"files":files})
	}else{
		context.Writer.WriteString(err.Error())
	}

}
func Html() {
	router := gin.Default()
	//gin.SetMode(gin.ReleaseMode)
	router.LoadHTMLGlob("templates/*")
	router.GET("/", index)
	router.GET("/lyear_main.html", main)
	router.GET("/disk.html", disk)
	router.StaticFS("/css", http.Dir("static/css"))
	router.StaticFS("/js", http.Dir("static/js"))
	router.StaticFS("/images", http.Dir("static/images"))
	router.StaticFS("/fonts", http.Dir("static/fonts"))
	router.Run(":6363")
}
