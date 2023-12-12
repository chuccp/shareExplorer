package io

import (
	"fmt"
	"github.com/chuccp/kuic/util"
	"github.com/juju/ratelimit"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

type FileInfo struct {
	IsDir      bool   `json:"isDir"`
	IsDisk     bool   `json:"isDisk"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	ModifyTime int64  `json:"modifyTime"`
}
type File struct {
	path   string
	normal string
	file   *os.File
	isDir  bool
}

func (f *File) List() ([]*FileInfo, error) {
	dirs, err := os.ReadDir(f.normal)
	if err != nil {
		return nil, err
	}
	var files = make([]*FileInfo, 0)
	for _, dir := range dirs {
		info, err := dir.Info()
		if err == nil {
			path := filepath.Join(f.normal, info.Name())
			IsHidden, err := IsHiddenFile(path)
			if err != nil || IsHidden {
				continue
			}
			files = append(files, &FileInfo{IsDir: dir.IsDir(), Name: info.Name(), Path: path, Size: info.Size(), ModifyTime: info.ModTime().UnixMilli()})
		}
	}
	return files, nil
}
func (f *File) ListDir() ([]*FileInfo, error) {
	dirs, err := os.ReadDir(f.normal)
	if err != nil {
		return nil, err
	}
	var files = make([]*FileInfo, 0)
	for _, dir := range dirs {
		info, err := dir.Info()
		if err == nil && dir.IsDir() {
			path := filepath.Join(f.normal, info.Name())
			IsHidden, err := IsHiddenFile(path)
			if err != nil || IsHidden {
				continue
			}
			files = append(files, &FileInfo{IsDir: dir.IsDir(), Name: info.Name(), Path: path, Size: info.Size(), ModifyTime: info.ModTime().UnixMilli()})
		}
	}
	return files, nil
}
func (f *File) Close() error {
	return f.file.Close()
}
func OpenFile(path string) (*File, error) {
	normal, err := filepath.Abs(path)
	if err != nil {
		fmt.Errorf("openfile %s", err)
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &File{normal: normal, path: path, file: file}, nil
}
func ReadFile(path string) ([]byte, error) {
	file, err := util.NewFile(path)
	if err != nil {
		fmt.Errorf("readfile %s", err)
		return nil, err
	}
	all, err := file.ReadAll()
	if err != nil {
		file.Close()
		return nil, err
	} else {
		file.Close()
		return all, nil
	}
}

func ReadChildrenDir(path string) ([]*FileInfo, error) {
	absolute := filepath.Join(path)
	file, err := OpenFile(absolute)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	list, err := file.ListDir()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func ReadRootPath() ([]*FileInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}
	if info.OS == "windows" {
		paths := make([]string, 0)
		var files = make([]*FileInfo, 0)
		partitions, err := disk.Partitions(false)
		if err != nil {
			return nil, err
		}
		for _, partition := range partitions {
			paths = append(paths, partition.Device+"\\")
		}
		for _, dir := range paths {
			info, err := os.Open(dir)
			if err == nil {
				stat, err := info.Stat()
				if err == nil {
					files = append(files, &FileInfo{IsDir: stat.IsDir(), IsDisk: true, Name: info.Name(), Path: info.Name(), Size: stat.Size(), ModifyTime: stat.ModTime().UnixMilli()})
				}
			}
			info.Close()
		}
		return files, nil
	} else {
		dirs, err := os.ReadDir("/")
		if err != nil {
			fmt.Errorf("readdir %s", err)
			return nil, err
		}
		var files = make([]*FileInfo, 0)
		for _, dir := range dirs {
			info, err := dir.Info()
			if err == nil {
				path := filepath.Join("/", info.Name())
				files = append(files, &FileInfo{IsDir: dir.IsDir(), IsDisk: false, Name: info.Name(), Path: path, Size: info.Size(), ModifyTime: info.ModTime().UnixMilli()})
			}
		}
		return files, nil
	}
}
func IsHiddenFile(filename string) (bool, error) {
	basename := filepath.Base(filename)
	if runtime.GOOS == "windows" {
		if basename[0:1] == "." {
			return true, nil
		}
		pointer, err := syscall.UTF16PtrFromString(filename)
		if err != nil {
			fmt.Errorf("IsHiddenFile %s", err)
			return false, err
		}
		attributes, err := syscall.GetFileAttributes(pointer)
		if err != nil {
			fmt.Errorf("IsHiddenFile %s", err)
			return false, err
		}
		return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
	} else {
		if basename[0:1] == "." {
			return true, nil
		}
	}
	return false, nil
}

func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	var bucket = ratelimit.NewBucketWithRate(100_000, 100)
	reader := ratelimit.Reader(src, bucket)
	return io.Copy(dst, reader)
}
