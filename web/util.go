package web

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}
func SaveUploadedFile2(src io.Reader, dst string, seq int) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if seq == 0 {
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}
	out, err := os.OpenFile(dst, flag, 0666)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}
