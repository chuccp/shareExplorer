package io

import (
	"os"
	"path/filepath"
	"strings"
)

type FileManage struct {
	base string
}

func (fm *FileManage) Children(path string) ([]*FileInfo, error) {
	path = strings.ReplaceAll(path, "\\", "/")
	absolute := filepath.Join(fm.base, path)
	file, err := OpenFile(absolute)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	list, err := file.List()
	for _, info := range list {
		info.Path, err = filepath.Rel(fm.base, info.Path)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (fm *FileManage) CreateNewFolder(path string, fileName string) error {
	path = strings.ReplaceAll(path, "\\", "/")
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		path = path[1:]
	}
	absolute := filepath.Join(fm.base, path, fileName)
	return os.MkdirAll(absolute, os.ModePerm)
}

func (fm *FileManage) Absolute(path string, fileName string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		path = path[1:]
	}
	absolute := filepath.Join(fm.base, path, fileName)
	return absolute
}

func CreateFileManage(root string) *FileManage {
	root = strings.ReplaceAll(root, "\\", "/")
	return &FileManage{base: root}
}
