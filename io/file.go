package io

import (
	"github.com/chuccp/kuic/util"
	"os"
	"path/filepath"
)

type FileInfo struct {
	IsDir      bool   `json:"isDir"`
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
			files = append(files, &FileInfo{IsDir: dir.IsDir(), Name: info.Name(), Path: path, Size: info.Size(), ModifyTime: info.ModTime().UnixMilli()})
		}
	}
	return files, nil
}
func OpenFile(path string) (*File, error) {
	normal, err := filepath.Abs(path)
	if err != nil {
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
