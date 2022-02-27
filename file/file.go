package file

import (
	"bufio"
	"errors"
	"github.com/yusufpapurcu/wmi"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
)

type File struct {
	normal string
	file   *os.File
	isDir  bool
	isDisk bool
}

func (f *File) Abs() string {
	return f.normal
}
func (f *File) Parent() string {
	return filepath.Dir(f.normal)
}
func (f *File) Name() string {
	if f.isDisk {
		return f.normal[0 : len(f.normal)-1]
	}
	return filepath.Base(f.normal)
}
func (f *File) ParentFile() (*File, error) {
	return NewFile(f.Parent())
}
func (f *File) open() error {
	if f.file == nil {
		file, err := os.Open(f.normal)
		if err != nil {
			return err
		}
		f.file = file
	}
	return nil
}
func (f *File) OpenAppendOrCreate() error {
	err := f.mkParent()
	if err != nil {
		return err
	}
	file1, err := os.OpenFile(f.normal, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	f.file = file1
	return err
}
func (f *File) mkParent() error {
	f, err := f.ParentFile()
	if err != nil {
		return err
	}
	err2 := f.MkDirs()
	return err2
}
func (f *File) OpenOrCreate() error {
	err := f.mkParent()
	if err != nil {
		return err
	}
	file1, err := os.OpenFile(f.normal, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	f.file = file1
	f.isDir = false
	return err
}

func (f *File) Exists() (flag bool, err error) {
	err = f.open()
	if err != nil {
		if strings.Contains(err.Error(), "cannot find the file") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f *File) MkDirs() error {
	err2 := os.MkdirAll(f.Abs(), 0666)
	return err2
}
func (f *File) OpenWrite() (*bufio.Writer, error) {
	err := f.OpenOrCreate()
	if err != nil {
		return nil, err
	}
	return bufio.NewWriter(f.file), nil
}
func (f *File) OpenAppendWrite() (*bufio.Writer, error) {
	err := f.OpenAppendOrCreate()
	if err != nil {
		return nil, err
	}
	return bufio.NewWriter(f.file), nil
}
func (f *File) Truncate() error {
	err := f.open()
	if err == nil {
		return f.file.Truncate(0)
	} else {
		return err
	}
}
func (f *File) Read() ([]byte,bool, error) {
	flag,err := f.Exists()
	if !flag{
		return nil, false, err
	}
	var allData = make([]byte, 0)
	var reader = bufio.NewReader(f.file)
	for {
		data:=make([]byte,1024)
		num, err := reader.Read(data)

		if err!=nil{
			if strings.Contains(err.Error(),"EOF"){
				break
			}
			return nil,true,err
		}
		if num==0{
			break
		}
		if err==nil{
			allData = append(allData, data[0:num]...)
		}
	}
	return allData, true,nil
}

func (f *File) WriteBytes(data []byte) error {
	if f.isDir {
		return errors.New(f.normal + " " + syscall.EISDIR.Error())
	}
	bw, err := f.OpenWrite()
	f.file.Truncate(0)
	if err != nil {
		return err
	}
	//os.Truncate(name, size)
	_, err = bw.Write(data)
	if err != nil {
		return err
	}
	return bw.Flush()
}

func (f *File) WriteAppendBytes(data []byte) error {
	if f.isDir {
		return errors.New(f.normal + " " + syscall.EISDIR.Error())
	}
	bw, err := f.OpenAppendWrite()
	if err != nil {
		return err
	}
	_, err = bw.Write(data)
	if err != nil {
		return err
	}
	return bw.Flush()
}
func (f *File) List() ([]*File, error) {
	dirs, err := os.ReadDir(f.normal)
	if err != nil {
		return nil, err
	}
	var files = make([]*File, 0)
	for _, dir := range dirs {
		file, err3 := NewFile(dir.Name())
		if err3 == nil {
			files = append(files, file)
		}
	}
	return files, err
}
func (f *File) Child(path string) (*File, error) {
	return NewFile(filepath.Join(f.normal, path))
}
func (f *File) IsDir() bool {
	return f.isDir
}
func (f *File) IsDisk() bool {
	return f.isDisk
}
func NewFile(path string) (*File, error) {
	normal, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fa, _ := regexp.MatchString(`^[a-zA-Z]:\\$`, path)
	if fa {
		return &File{normal: normal, isDir: true, isDisk: true}, nil
	}
	fileInfo, err := os.Stat(normal)
	if err != nil {
		if strings.Contains(err.Error(), "cannot find the file")|| strings.Contains(err.Error(), "cannot find the path"){
			return &File{normal: normal}, nil
		}
		return nil, err
	}
	return &File{normal: normal, isDir: fileInfo.IsDir()}, nil
}
func GetRootPath() ([]*File, error) {
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

func getWindowsRootPath() ([]*File, error) {
	var storageInfo []storageInfo
	err := wmi.Query("Select * from Win32_LogicalDisk", &storageInfo)
	files := make([]*File, 0)
	if err == nil {
		for _, v := range storageInfo {
			f, err := NewFile(v.Name + (string)(filepath.Separator))
			if err == nil {
				files = append(files, f)
			}
		}
	}
	return files, err
}
func getOtherRootPath() ([]*File, error) {
	dirs, err := os.ReadDir("/")
	if err == nil {
		files := make([]*File, 0)
		for _, v := range dirs {
			fi, err8 := NewFile(v.Name())
			if err8 == nil {
				files = append(files, fi)
			}
		}
		return files, nil
	}
	return nil, err
}
