package file

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/io"
	"github.com/chuccp/shareExplorer/web"
	"os"
)

type Server struct {
	fileManage *io.FileManage
}

func (s *Server) index(req *web.Request) (any, error) {
	return "首页", nil
}

func (s *Server) files(req *web.Request) (any, error) {
	Path := req.FormValue("Path")
	if len(Path) > 0 {
		child, err := s.fileManage.Children(Path)
		if err != nil {
			return nil, err
		} else {
			return child, err
		}
	}
	return nil, os.ErrNotExist
}
func (s *Server) Upload(req *web.Request) (any, error) {
	file, err := req.FormFile("file")
	if err != nil {
		return nil, err
	}
	Path := req.FormValue("Path")
	absolute := s.fileManage.Absolute(Path, file.Filename)
	err = web.SaveUploadedFile(file, absolute)
	if err != nil {
		return nil, err
	}
	return file.Filename, err
}

func (s *Server) GetName() string {
	return "file"
}

func (s *Server) Init(context *core.Context) {
	s.fileManage = io.CreateFileManage("C:\\Users\\cooge\\Documents")
	context.Get("/file/index", s.index)
	context.Get("/file/files", s.files)
	context.Get("/file/upload", s.Upload)
}
