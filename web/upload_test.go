package web

import (
	"bytes"
	"os"
	"testing"
)

func SplitCopyFile(src string, dst string, num int) ([]*TempUpload, error) {
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

func TestName(t *testing.T) {

	tempUploads, err := SplitCopyFile("C:\\Users\\cooge\\Downloads\\apache-tomcat-8.5.97-windows-x64.zip", "C:\\Users\\cooge\\Downloads\\apache-tomcat-8.5.97-windows-x64(2).zip", 10)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(tempUploads)

	for _, v := range tempUploads {
		err := v.SaveUploaded()
		if err != nil {
			t.Log(err)
			return
		}
	}

}
