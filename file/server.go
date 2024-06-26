package file

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/io"
	"github.com/chuccp/shareExplorer/web"
	"go.uber.org/zap"
	"os"
	"path"
)

type Server struct {
	context     *core.Context
	webdavStore *webDavStore
}

func (s *Server) index(req *web.Request) (any, error) {
	return "首页", nil
}

func (s *Server) files(req *web.Request) (any, error) {
	Path := req.FormValue("Path")
	RootPath := req.FormValue("RootPath")
	if len(Path) > 0 && len(RootPath) > 0 {
		rootPath, err := s.getRealRootPath(RootPath)
		if err != nil {
			return nil, err
		}
		fileManage := io.CreateFileManage(rootPath)
		children, err := fileManage.Children(Path)
		if err != nil {
			return nil, err
		} else {
			return web.ResponseOK(children), err
		}
	}
	return nil, os.ErrNotExist
}

func (s *Server) getRealRootPath(rootPath string) (string, error) {
	query, err := s.context.GetDB().GetPathModel().Query(rootPath)
	if err != nil {
		return "", err
	}
	return query.Path, nil
}

func (s *Server) download(req *web.Request) (any, error) {
	Path := req.FormValue("path")
	RootPath := req.FormValue("rootPath")
	if len(Path) > 0 && len(RootPath) > 0 {
		rootPath, err := s.getRealRootPath(RootPath)
		if err != nil {
			return nil, err
		}
		file := path.Join(rootPath, Path)
		return web.ResponseFile(file), nil
	}
	return nil, os.ErrNotExist
}

func (s *Server) rename(req *web.Request) (any, error) {

	var rename entity.Rename
	v, err := req.BodyJson(&rename)
	if err != nil {
		s.context.GetLog().Error("addUser", zap.Error(err), zap.ByteString("body", v))
		return nil, os.ErrNotExist
	}
	if len(rename.Path) > 0 && len(rename.RootPath) > 0 {
		rootPath, err := s.getRealRootPath(rename.RootPath)
		if err != nil {
			return nil, err
		}
		fileDir := path.Join(rootPath, rename.Path)
		oldName := path.Join(fileDir, rename.OldName)
		newFile := path.Join(fileDir, rename.NewName)
		err = os.Rename(oldName, newFile)
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
		rootPath, err := s.getRealRootPath(RootPath)
		if err != nil {

		}
		file := path.Join(rootPath, Path)
		println(file)
		err = os.Remove(file)
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
		rootPath, err := s.getRealRootPath(rootPath)
		if err != nil {
			return nil, err
		}
		fileManage := io.CreateFileManage(rootPath)
		absolute := fileManage.Absolute(path, name)
		tempUpload := web.NewTempUpload(reader, absolute, seq, count, size, total)
		err = tempUpload.SaveUploaded()
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
	rootPath, err := s.getRealRootPath(rootPath)
	if err != nil {
		return nil, err
	}
	fileManage := io.CreateFileManage(rootPath)
	absolute := fileManage.Absolute(path, name)
	err = web.CancelTempUpload(absolute)
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
	v, err := req.BodyJson(&folder)
	if err != nil {
		s.context.GetLog().Error("addUser", zap.Error(err), zap.ByteString("body", v))
		return nil, err
	}
	if len(folder.Path) > 0 && len(folder.Folder) > 0 && len(folder.RootPath) > 0 {
		rootPath, err := s.getRealRootPath(folder.RootPath)
		if err != nil {
			return nil, err
		}
		fileManage := io.CreateFileManage(rootPath)
		err = fileManage.CreateNewFolder(folder.Path, folder.Folder)
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
	return web.ResponseOK(pathInfo), nil
}
func (s *Server) paths(req *web.Request) (any, error) {
	Path := req.FormValue("Path")
	if len(Path) > 0 {
		pathInfo, err := io.ReadChildrenDir(Path)
		if err != nil {
			return nil, err
		} else {
			return web.ResponseOK(pathInfo), err
		}
	}
	return nil, os.ErrNotExist
}

func (s *Server) dav(req *web.Request) (any, error) {

	s.context.GetLog().Debug("dav", zap.String("RequestURI", req.GetRawRequest().RequestURI))

	s.webdavStore.getWebdav(req.GetAuthUsername()).ServeHTTP(req.GetResponseWriter(), req.GetRawRequest())
	return nil, nil
}

func (s *Server) queryAllPath(req *web.Request) (any, error) {
	list, num, err := s.context.GetDB().GetPathModel().QueryPage(1, 100)
	if err != nil {
		return nil, err
	}
	for _, p := range list {
		p.Path = p.Name
	}
	pageAble := &web.PageAble{Total: num, List: list}
	return web.ResponseOK(pageAble), nil
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	s.webdavStore = newWebDavStore(context, "/dav")
	context.AnyRemoteAuth("/dav/*name", s.dav)
	context.GetRemoteAuth("/file/queryAllPath", s.queryAllPath)
	context.GetRemoteAuth("/file/root", s.root)
	context.GetRemoteAuth("/file/paths", s.paths)
	context.GetRemoteAuth("/file/index", s.index)
	context.GetRemoteAuth("/file/download", s.download)
	context.PostRemoteAuth("/file/rename", s.rename)
	context.GetRemoteAuth("/file/delete", s.delete)
	context.GetRemoteAuth("/file/files", s.files)
	context.PostRemoteAuth("/file/upload", s.upload)
	context.GetRemoteAuth("/file/cancelUpload", s.cancelUpload)
	context.PostRemoteAuth("/file/createNewFolder", s.createNewFolder)
}
