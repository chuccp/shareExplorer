package web

import (
	"github.com/chuccp/kuic/cert"
	khttp "github.com/chuccp/kuic/http"
	"github.com/chuccp/shareExplorer/io"
	"github.com/gin-gonic/gin"
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

func (s *Server) Init() {
	s.fileManage = io.CreateFileManage("C:\\Users\\cooge\\Downloads")
	s.engine.GET("/index", s.Index)
	s.engine.GET("/files", s.Files)
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
	server.Listen(certPath, keyPath, s.engine)
	return nil
}

func NewServer() *Server {

	return &Server{engine: gin.Default()}
}
