package web

import (
	"io"
	"mime/multipart"
	"os"
	"path"
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
func SaveData(file []byte, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.Write(file)
	return err
}
func ReadAllUploadedFile(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	return io.ReadAll(src)
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
func SaveUploadedFileTemp(src io.Reader, dst string, seq int, count int64, size int, total int) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if seq == 0 {
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}
	dir, file := path.Split(dst)
	out, err := os.OpenFile(path.Join(dir, "."+file), flag, 0666)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}
