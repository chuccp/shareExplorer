package io

import (
	"os"
	"path/filepath"
)

type FileInfo struct {
	IsDir bool   `json:"isDir"`
	Name  string `json:"name"`
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
			files = append(files, &FileInfo{IsDir: dir.IsDir(), Name: info.Name()})
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
