package file

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/io"
	"github.com/chuccp/shareExplorer/web"
	"os"
	"path"
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

func (s *Server) download(req *web.Request) (any, error) {
	Path := req.FormValue("path")
	RootPath := req.FormValue("rootPath")
	if len(Path) > 0 && len(RootPath) > 0 {
		file := path.Join(RootPath, Path)
		return web.ResponseFile(file), nil
	}
	return nil, os.ErrNotExist
}

func (s *Server) rename(req *web.Request) (any, error) {

	var rename entity.Rename
	err := req.BodyJson(&rename)
	if err != nil {
		return nil, os.ErrNotExist
	}
	if len(rename.Path) > 0 && len(rename.RootPath) > 0 {
		fileDir := path.Join(rename.RootPath, rename.Path)
		oldName := path.Join(fileDir, rename.OldName)
		newFile := path.Join(fileDir, rename.NewName)
		err := os.Rename(oldName, newFile)
		if err != nil {
			return nil, err
		}
		return web.ResponseOK("ok"), nil
	}
	return nil, os.ErrNotExist
}
func (s *Server) delete(req *web.Request) (any, error) {
	Path := req.FormValue("path")
	RootPath := req.FormValue("rootPath")
	if len(Path) > 0 && len(RootPath) > 0 {
		file := path.Join(RootPath, Path)
		println(file)
		err := os.Remove(file)
		if err != nil {
			return nil, err
		}
		return web.ResponseOK("ok"), nil
	}
	return nil, os.ErrNotExist
}
func (s *Server) upload(req *web.Request) (any, error) {
	reader := req.GetRawRequest().Body
	path := req.FormValue("path")
	rootPath := req.FormValue("rootPath")
	name := req.FormValue("name")
	seq := req.FormIntValue("seq")
	count := req.FormIntValue("count")
	size := req.FormInt64Value("size")
	total := req.FormInt64Value("total")
	if len(path) > 0 && len(rootPath) > 0 && len(name) > 0 && size > 0 && total > 0 && count > 0 {
		fileManage := io.CreateFileManage(rootPath)
		absolute := fileManage.Absolute(path, name)
		tempUpload := web.NewTempUpload(reader, absolute, seq, count, size, total)
		err := tempUpload.SaveUploaded()
		if err != nil {
			return nil, err
		}
		return web.ResponseOK(name), err
	}
	return nil, os.ErrNotExist
}
func (s *Server) cancelUpload(req *web.Request) (any, error) {
	path := req.FormValue("path")
	rootPath := req.FormValue("rootPath")
	name := req.FormValue("name")
	fileManage := io.CreateFileManage(rootPath)
	absolute := fileManage.Absolute(path, name)
	err := web.CancelTempUpload(absolute)
	if err != nil {
		return nil, err
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
	context.GetRemote("/file/download", s.download)
	context.PostRemote("/file/rename", s.rename)
	context.GetRemote("/file/delete", s.delete)
	context.GetRemote("/file/files", s.files)
	context.PostRemote("/file/upload", s.upload)
	context.GetRemote("/file/cancelUpload", s.cancelUpload)
	context.PostRemote("/file/createNewFolder", s.createNewFolder)
}
