package web

import "path/filepath"

type File struct {
	path string
}

func (f *File) GetPath() string {
	return f.path
}
func (f *File) GetFilename() string {
	return filepath.Base(f.path)
}
func ResponseFile(path string) *File {
	return &File{path: path}
}
