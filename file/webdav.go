package file

import (
	"context"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/util"
	"golang.org/x/net/webdav"
	"io/fs"
	"log"
	"os"
	"path"
	"sync"
)

var contextInfoKey = "_contextInfo_"

type contextInfo struct {
	Username string
}

type webDavStore struct {
	webdavMap map[string]*webdav.Handler
	lock      sync.Mutex
	context   *core.Context
	prefix    string
}

func newWebDavStore(context *core.Context, prefix string) *webDavStore {
	return &webDavStore{webdavMap: make(map[string]*webdav.Handler), context: context, prefix: prefix}
}

func (s *webDavStore) getWebdav(username string) *webdav.Handler {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, ok := s.webdavMap[username]
	if ok {
		return v
	}
	wb := &webdav.Handler{
		FileSystem: NewDavFileSystem(s.context, username, s.prefix),
		LockSystem: webdav.NewMemLS(),
		Prefix:     s.prefix,
	}
	s.webdavMap[username] = wb
	return wb
}

type webdavPool struct {
}

type DavFileSystem struct {
	username string
	context  *core.Context
	prefix   string
}

type webPath struct {
	rawPath string
	path    string
	paths   []string
}
type davFile struct {
	webdav.File
	context *core.Context
}

func newDavFile(file webdav.File, context *core.Context) *davFile {
	return &davFile{File: file, context: context}
}

type fileInfo struct {
	fs.FileInfo
	name string
}

func (i *fileInfo) Name() string {
	return i.name
}
func newFileInfo(f fs.FileInfo, name string) fs.FileInfo {
	return &fileInfo{FileInfo: f, name: name}
}
func (file *davFile) Readdir(count int) ([]fs.FileInfo, error) {

	paths, err := file.context.GetDB().GetPathModel().QueryAll()
	if err != nil {
		return nil, err
	}
	files := make([]fs.FileInfo, 0)
	for _, path := range paths {
		log.Println(path.Path)
		open, err := os.Open(path.Path)
		if err != nil {
			continue
		} else {
			stat, err := open.Stat()
			if err != nil {
				continue
			} else {
				files = append(files, newFileInfo(stat, path.Name))
			}
		}
	}
	log.Println(len(files))
	return files, nil
}

func (wp *webPath) isRoot() bool {
	return len(wp.paths) == 0
}
func (wp *webPath) Path() string {

	if wp.isRoot() {
		return ""
	}

	return path.Join(wp.paths[1:]...)
}
func (wp *webPath) name() string {
	if wp.isRoot() {
		return ""
	}
	return wp.paths[0]
}
func newWebPath(path_ string) *webPath {
	log.Println("path", path_)
	ps := util.SplitPath(path_)
	return &webPath{rawPath: path_, paths: ps}
}

func NewDavFileSystem(context *core.Context, username string, prefix string) *DavFileSystem {
	return &DavFileSystem{username: username, context: context, prefix: prefix}
}

func getValue(ctx context.Context) *contextInfo {
	return ctx.Value(contextInfoKey).(*contextInfo)
}
func (d *DavFileSystem) getDir(ctx context.Context, name string) (webdav.Dir, *webPath, error) {
	webPath := newWebPath(name)
	if webPath.isRoot() {
		dir := webdav.Dir("")
		return dir, webPath, nil
	}
	query, err := d.context.GetDB().GetPathModel().Query(webPath.name())
	if err != nil {
		return "", webPath, err
	}
	dir := webdav.Dir(query.Path)
	return dir, webPath, nil
}

func (d *DavFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	log.Println("Mkdir", name)
	dir, webPath, err := d.getDir(ctx, name)
	if err != nil {
		return err
	}
	return dir.Mkdir(ctx, webPath.Path(), perm)
}
func (d *DavFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	log.Println("OpenFile", name)
	dir, webPath, err := d.getDir(ctx, name)
	if err != nil {
		return nil, err
	}
	if webPath.isRoot() {
		file, err := dir.OpenFile(ctx, "", flag, perm)
		if err != nil {
			return nil, err
		} else {
			return newDavFile(file, d.context), nil
		}
	}
	return dir.OpenFile(ctx, webPath.Path(), flag, perm)
}
func (d *DavFileSystem) RemoveAll(ctx context.Context, name string) error {
	log.Println("RemoveAll", name)
	dir, webPath, err := d.getDir(ctx, name)
	if err != nil {
		return err
	}
	return dir.RemoveAll(ctx, webPath.Path())
}
func (d *DavFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	log.Println("Rename", oldName)
	dir, webPath, err := d.getDir(ctx, oldName)
	if err != nil {
		return err
	}
	newWebPath := newWebPath(newName)
	return dir.Rename(ctx, webPath.Path(), newWebPath.Path())
}
func (d *DavFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	log.Println("Stat", name)
	dir, webPath, err := d.getDir(ctx, name)
	if err != nil {
		return nil, err
	}
	return dir.Stat(ctx, webPath.Path())
}
