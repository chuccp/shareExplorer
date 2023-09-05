package web

import (
	"github.com/chuccp/kuic/cert"
	khttp "github.com/chuccp/kuic/http"
	"github.com/chuccp/shareExplorer/io"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

type Server struct {
	engine     *gin.Engine
	fileManage *io.FileManage
}

func (s *Server) Index(context *gin.Context) {
	context.Writer.Write([]byte("Index"))
}

func (s *Server) Files(context *gin.Context) {

	Path := context.Request.FormValue("Path")
	if len(Path) > 0 {
		child, err := s.fileManage.Children(Path)
		if err != nil {
			context.AbortWithError(500, err)
		} else {
			context.AbortWithStatusJSON(200, child)
		}
	} else {
		context.AbortWithError(404, os.ErrNotExist)
	}
}
func (s *Server) Upload(context *gin.Context) {

	file, err := context.FormFile("file")
	if err != nil {
		context.AbortWithError(500, err)
		return
	}
	Path := context.Request.FormValue("Path")
	absolute := s.fileManage.Absolute(Path, file.Filename)
	err = context.SaveUploadedFile(file, absolute)
	if err != nil {
		context.AbortWithError(500, err)
		return
	}
	context.String(http.StatusOK, file.Filename)
}
func (s *Server) Init() {
	s.fileManage = io.CreateFileManage("C:/Users/cooge/Pictures/")
	s.engine.GET("/index", s.Index)
	s.engine.GET("/files", s.Files)
	s.engine.POST("/upload", s.Upload)
}
func (s *Server) Start() error {
	keyPath := "keyPath.PEM"
	certPath := "certPath.PEM"
	server, err := khttp.CreateServer("0.0.0.0:2156")
	if err != nil {
		return err
	}
	err = cert.CreateOrReadCert(keyPath, certPath)
	if err != nil {
		return err
	}

	server.ListenAndServe(certPath, keyPath, s.engine)
	return nil
}

func NewServer() *Server {
	server := &Server{engine: gin.Default()}
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 允许的域名列表，可以使用 * 来允许所有域名
	server.engine.Use(cors.New(config))
	return server
}
