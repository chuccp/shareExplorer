package io

import (
	"path/filepath"
)

func File2FileInfo(base string, file *File) (*FileInfo, error) {
	relative, err := filepath.Rel(base, file.normal)
	if err != nil {
		return nil, err
	}
	return &FileInfo{IsDir: file.isDir, Relative: relative}, nil
}

type FileManage struct {
	base string
}

func (fm *FileManage) Children(path string) ([]*FileInfo, error) {
	absolute := filepath.Join(fm.base, path)
	file, err := OpenFile(absolute)
	if err != nil {
		return nil, err
	}
	list, err := file.List(fm.base)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func CreateFileManage(root string) *FileManage {
	return &FileManage{base: root}
}
