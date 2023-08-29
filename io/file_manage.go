package io

import (
	"path/filepath"
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

func CreateFileManage(root string) *FileManage {
	return &FileManage{base: root}
}
