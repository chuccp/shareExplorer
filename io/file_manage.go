package io

import (
	"path/filepath"
	"strings"
)

type FileManage struct {
	base string
}

func (fm *FileManage) Children(path string) ([]*FileInfo, error) {
	absolute := filepath.Join(fm.base, path)
	file, err := OpenFile(absolute)
	if err != nil {
		return nil, err
	}
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

func (fm *FileManage) Absolute(path string, fileName string) string {
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		path = path[1:]
	}
	absolute := filepath.Join(fm.base, path, fileName)
	return absolute
}

func CreateFileManage(root string) *FileManage {
	return &FileManage{base: root}
}
