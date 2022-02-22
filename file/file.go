package file

import (
	"github.com/yusufpapurcu/wmi"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetRootPath() ([]*File, error) {
	fs, err := getRootPath()
	if err == nil {
		files := make([]*File, len(fs))
		for i, _ := range files {
			files[i] = &File{file: fs[i], isDir: true,isDisk: true}
		}
		return files, err
	}
	return nil, err
}

func getRootPath() ([]*os.File, error) {
	if runtime.GOOS == "windows" {
		return getWindowsRootPath()
	}
	return getOtherRootPath()
}

type storageInfo struct {
	Name       string
	Size       uint64
	FreeSpace  uint64
	FileSystem string
}
type DirEntry struct {
	Parent string
	dir    os.DirEntry
}

func (dir *DirEntry) ToFile() (*File, error) {
	f, err := os.Open(dir.Parent + dir.dir.Name())
	if err == nil {
		return &File{file: f, Parent: dir.Parent, isDir: dir.IsDir()}, err
	}
	return nil, err
}
func (dir *DirEntry) IsDir() bool {
	return dir.dir.IsDir()
}




type File struct {
	Parent string
	file   *os.File
	isDir  bool
	isDisk bool
}
func NewFilePath(parent string,relativePath string)(*File, error){
	return NewFile(filepath.Join(parent,relativePath))
}
func NewFile(path string) (*File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	file2 := &File{file: file, isDir: fi.IsDir()}
	path = fi.Name()
	num := strings.Count(path, string(filepath.Separator))
	if num > 1 {
		last := strings.LastIndex(path, string(filepath.Separator))
		parentPath := path[:last]
		file2.Parent = parentPath
	}
	return file2, nil
}

func (fi *File) Abs() string {
	return fi.file.Name()
}
func (fi *File) GetParent() (*File, error) {
	return NewFile(fi.Parent)
}
func (fi *File) abs() string {
	if fi.isDir {
		sq := string(filepath.Separator)
		if strings.HasSuffix(fi.file.Name(), sq) {
			return fi.file.Name()
		}
		return fi.file.Name() + sq
	}
	return fi.file.Name()
}
func (fi *File) IsDir() bool {
	st, _ := fi.file.Stat()
	return st.IsDir()
}
func (fi *File) ListAllFile() ([]*File, error) {

	return fi.ListFile(0)
}
func (fi *File) ListFile(n int) ([]*File, error) {
	dirs, err := fi.List(n)
	if err != nil {
		return nil, err
	}
	files := make([]*File, 0)
	for _, dir := range dirs {
		f, err5 := dir.ToFile()
		if err5 == nil {
			files = append(files, f)
		}
	}
	return files, err
}
func (fi *File) List(n int) ([]*DirEntry, error) {
	dirs, err := fi.file.ReadDir(n)

	if err != nil {
		return nil, err
	}
	vDirs := make([]*DirEntry, 0)
	for _, f := range dirs {

		vDirs = append(vDirs, &DirEntry{dir:f, Parent: fi.abs()})
	}
	return vDirs, err
}

func (fi *File) Name() string {
	name := fi.file.Name()
	log.Println( "====",name)
	index := strings.LastIndexByte(fi.file.Name(), filepath.Separator)
	if index > -1 {
		if fi.isDisk{
			return name[0:index]
		}else{
			return name[index+1:]
		}
	}
	return name
}
func (fi *File) IsDisk()bool{
	return fi.isDisk
}
func getOtherRootPath() ([]*os.File, error) {
	f, err := os.Open("/")
	files := make([]*os.File, 0)
	if err == nil {
		files = append(files, f)
	}
	return files, err
}
func getWindowsRootPath() ([]*os.File, error) {
	var storageInfo []storageInfo
	err := wmi.Query("Select * from Win32_LogicalDisk", &storageInfo)
	files := make([]*os.File, 0)
	if err == nil {
		for _, v := range storageInfo {
			f, err := os.Open(v.Name + string(filepath.Separator))
			if err == nil {
				files = append(files, f)
			}
		}
	}
	return files, err
}
