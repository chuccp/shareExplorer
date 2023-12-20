package web

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func SplitFile(src string, dst string, num int) ([]*TempUpload, error) {
	srcFile, err := os.OpenFile(src, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()
	stat, err := srcFile.Stat()
	if err != nil {
		return nil, err
	}
	total := stat.Size()
	nums := splitNumber(total, int64(num))
	tempUploads := make([]*TempUpload, len(nums))

	data, err := os.ReadFile(src)
	if err != nil {
		return nil, err
	}

	size := int64(0)

	for i, _ := range tempUploads {
		buff := bytes.NewReader(data[size:(size + nums[i])])
		size = size + nums[i]
		tempUploads[i] = NewTempUpload(buff, dst, i, len(nums), nums[i], total)
	}
	return tempUploads, nil
}

type TempUpload struct {
	dst    string
	seq    int
	count  int
	size   int64
	total  int64
	src    io.Reader
	temp   string
	upload string
	dir    string
	file   string
}

var uploadSuffix = ".upload"
var tempSuffix = ".temp"

func NewTempUpload(src io.Reader, dst string, seq int, count int, size int64, total int64) *TempUpload {
	tu := &TempUpload{dst: dst, seq: seq, count: count, size: size, total: total, src: src}
	dst = strings.ReplaceAll(dst, "\\", "/")
	tu.dir, tu.file = path.Split(dst)
	tu.temp = tu.file + tempSuffix
	tu.upload = tu.file + uploadSuffix
	return tu
}
func (tempUpload *TempUpload) SaveUploaded() error {

	if tempUpload.seq != 0 {
		records, err := tempUpload.readRecord()
		if err != nil {
			return err
		}
		fa := tempUpload.validated(records)
		if !fa {
			return errors.New("文件错误")
		}
		err = tempUpload.writeUpload(records)
		if err != nil {
			return err
		}
		if tempUpload.seq == tempUpload.count-1 {
			tempUpload.finish()
		}
		return nil
	} else {
		err := tempUpload.writeUpload(nil)
		if err != nil {
			return err
		}

		return nil
	}

}
func (tempUpload *TempUpload) writeUpload(preTempUploads []*TempUpload) error {
	file, err := tempUpload.readFile()
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() > tempUpload.total {
		return errors.New("文件不对")
	}
	size := tempUpload.readSize(tempUpload.seq, preTempUploads)
	if preTempUploads != nil && tempUpload.seq == len(preTempUploads) {
		file.Seek(size, 0)
		_, err := io.Copy(file, tempUpload.src)
		if err != nil {
			return err
		}
		err = tempUpload.writeTempAppend()
		if err != nil {
			return err
		}
	} else {
		file.Seek(size, 0)
		_, err := io.Copy(file, tempUpload.src)
		if err != nil {
			return err
		}
		if tempUpload.seq == 0 {
			err := tempUpload.writeTempNew0()
			if err != nil {
				return err
			}
		} else {
			err := tempUpload.writeTempNew(preTempUploads[0:tempUpload.seq])
			if err != nil {
				return err
			}
		}

	}
	return nil
}
func (tempUpload *TempUpload) readSize(seq int, tempUploads []*TempUpload) int64 {
	if tempUpload.seq == 0 {
		return 0
	}
	var num int64 = 0
	for i, upload := range tempUploads {
		if i >= seq {
			break
		} else {
			num = num + upload.size
		}
	}
	return num
}

func (tempUpload *TempUpload) validated(preTempUploads []*TempUpload) bool {
	if tempUpload.seq == 0 {
		return true
	}
	temp := preTempUploads[0]
	if temp.total != tempUpload.total {
		return false
	}
	numbers := splitNumber(temp.total, int64(temp.count))
	if numbers[tempUpload.seq] != tempUpload.size {
		return false
	}
	if tempUpload.seq > len(preTempUploads) {
		return false
	}
	return true
}

func splitNumber(total int64, count int64) []int64 {
	var result []int64
	quotient := total / count
	remainder := total % count
	for i := int64(0); i < count; i++ {
		if i < remainder {
			result = append(result, quotient+int64(1))
		} else {
			result = append(result, quotient)
		}
	}
	return result
}

func (tempUpload *TempUpload) readRecord() ([]*TempUpload, error) {
	temp, err := tempUpload.readTemp()
	if err != nil {
		return nil, err
	}
	defer temp.Close()
	tempUploads := make([]*TempUpload, 0)
	if tempUpload.seq == 0 {
		return tempUploads, nil
	}
	scanner := bufio.NewScanner(temp)
	for scanner.Scan() {
		lastLine := scanner.Text()
		pre, err := tempUpload.deserialize(lastLine)
		if err != nil {
			return nil, err
		}
		tempUploads = append(tempUploads, pre)
	}
	return tempUploads, nil
}
func (tempUpload *TempUpload) deserialize(str string) (*TempUpload, error) {
	as := strings.Split(str, "_")
	var tmp TempUpload
	var err error
	if len(as) == 4 {
		tmp.seq, err = strconv.Atoi(as[0])
		if err != nil {
			return nil, err
		}
		tmp.count, err = strconv.Atoi(as[1])
		if err != nil {
			return nil, err
		}
		tmp.size, err = strconv.ParseInt(as[2], 10, 64)
		if err != nil {
			return nil, err
		}
		tmp.total, err = strconv.ParseInt(as[3], 10, 64)
		if err != nil {
			return nil, err
		}
		return &tmp, nil
	}

	return nil, errors.New("")
}
func (tempUpload *TempUpload) serialize() string {
	return strconv.Itoa(tempUpload.seq) + "_" + strconv.Itoa(tempUpload.count) + "_" + strconv.FormatInt(tempUpload.size, 10) + "_" + strconv.FormatInt(tempUpload.total, 10)
}
func (tempUpload *TempUpload) finish() error {
	uploadPath := path.Join(tempUpload.dir, strings.TrimSuffix(tempUpload.upload, uploadSuffix))
	uploadTempPath := path.Join(tempUpload.dir, tempUpload.upload)
	err := os.Rename(uploadTempPath, uploadPath)
	if err != nil {
		return err
	}
	err = os.Remove(path.Join(tempUpload.dir, tempUpload.temp))
	if err != nil {
		return err
	}
	return nil
}
func (tempUpload *TempUpload) readTemp() (temp *os.File, err error) {
	temp, err = os.OpenFile(path.Join(tempUpload.dir, tempUpload.temp), os.O_RDONLY, 0666)
	if err != nil {
		return
	}
	return
}
func (tempUpload *TempUpload) writeTempAppend() error {
	temp, err := os.OpenFile(path.Join(tempUpload.dir, tempUpload.temp), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer temp.Close()
	if tempUpload.seq == 0 {
		_, err := temp.WriteString(tempUpload.serialize())
		if err != nil {
			return err
		}
	} else {
		_, err := temp.WriteString("\n" + tempUpload.serialize())
		if err != nil {
			return err
		}
	}
	return nil
}

func (tempUpload *TempUpload) writeTempNew(tempUploads []*TempUpload) error {
	temp, err := os.OpenFile(path.Join(tempUpload.dir, tempUpload.temp), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer temp.Close()
	for i, upload := range tempUploads {
		if i == 0 {
			_, err := temp.WriteString(upload.serialize())
			if err != nil {
				return err
			}
		} else {
			_, err := temp.WriteString("\n" + upload.serialize())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (tempUpload *TempUpload) writeTempNew0() error {
	temp, err := os.OpenFile(path.Join(tempUpload.dir, tempUpload.temp), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer temp.Close()
	_, err = temp.WriteString(tempUpload.serialize())
	if err != nil {
		return err
	}
	return nil
}
func (tempUpload *TempUpload) readFile() (temp *os.File, err error) {
	flag := os.O_RDWR | os.O_CREATE
	temp, err = os.OpenFile(path.Join(tempUpload.dir, tempUpload.upload), flag, 0666)
	if err != nil {
		return
	}
	return
}
func (tempUpload *TempUpload) SaveUploadedFileTemp() error {
	if err := os.MkdirAll(filepath.Dir(tempUpload.dst), 0750); err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if tempUpload.seq == 0 {
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}
	dir, file := path.Split(tempUpload.dst)
	fileName := file + ".temp"
	out, err := os.OpenFile(path.Join(dir, fileName), flag, 0666)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, tempUpload.src)
	return err
}
