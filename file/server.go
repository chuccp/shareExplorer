package file

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/io"
	"github.com/chuccp/shareExplorer/web"
	"os"
)

type Server struct {
}

func (s *Server) index(req *web.Request) (any, error) {
	return "首页", nil
}

func (s *Server) files(req *web.Request) (any, error) {
	Path := req.FormValue("Path")
	RootPath := req.FormValue("RootPath")
	if len(Path) > 0 && len(RootPath) > 0 {
		fileManage := io.CreateFileManage(RootPath)
		child, err := fileManage.Children(Path)
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
	RootPath := req.FormValue("RootPath")
	if len(Path) > 0 && len(RootPath) > 0 {
		fileManage := io.CreateFileManage(RootPath)
		absolute := fileManage.Absolute(Path, file.Filename)
		err = web.SaveUploadedFile(file, absolute)
		if err != nil {
			return nil, err
		}
		return file.Filename, err
	}
	return nil, os.ErrNotExist

}

type NewFolder struct {
	Folder   string `json:"folder"`
	RootPath string `json:"rootPath"`
	Path     string `json:"path"`
}

func (s *Server) createNewFolder(req *web.Request) (any, error) {
	var folder NewFolder
	err := req.BodyJson(&folder)
	if err != nil {
		return nil, err
	}
	if len(folder.Path) > 0 && len(folder.Folder) > 0 && len(folder.RootPath) > 0 {
		fileManage := io.CreateFileManage(folder.RootPath)
		err := fileManage.CreateNewFolder(folder.Path, folder.Folder)
		if err != nil {
			return nil, err
		} else {
			return web.ResponseOK("ok"), nil
		}
	}
	return nil, os.ErrNotExist

}

func (s *Server) GetName() string {
	return "file"
}

func (s *Server) root(req *web.Request) (any, error) {
	pathInfo, err := io.ReadRootPath()
	if err != nil {
		return nil, err
	}
	return pathInfo, nil
}
func (s *Server) paths(req *web.Request) (any, error) {
	Path := req.FormValue("Path")
	if len(Path) > 0 {
		pathInfo, err := io.ReadChildrenDir(Path)
		if err != nil {
			return nil, err
		} else {
			return pathInfo, err
		}
	}
	return nil, os.ErrNotExist
}

func (s *Server) Init(context *core.Context) {
	context.GetRemote("/file/root", s.root)
	context.GetRemote("/file/paths", s.paths)
	context.GetRemote("/file/index", s.index)
	context.GetRemote("/file/files", s.files)
	context.PostRemote("/file/upload", s.Upload)
	context.PostRemote("/file/createNewFolder", s.createNewFolder)
}
