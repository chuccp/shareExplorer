package io

import (
	"log"
	"os"
	"testing"
)

func TestName(t *testing.T) {
	fileManage := CreateFileManage("C:\\Users\\cooge\\Downloads")

	children, err := fileManage.Children("/")
	if err != nil {
		t.Error(err)
		return
	}

	log.Println(len(children))
}
func TestRoot(t *testing.T) {

	dir, err := ReadChildrenDir("c:\\")
	t.Log(err)
	if err != nil {
		return
	}
	for _, info := range dir {
		println(info.Path)
	}

}

func TestCopy(t *testing.T) {

	file, err := os.Open("C:\\Users\\cooge\\Downloads\\11111.jpg")
	if err != nil {
		return
	}

	out, err := os.Create("C:\\Users\\cooge\\Downloads\\22222.jpg")
	if err != nil {
		return
	}
	_, err = Copy(out, file)
	if err != nil {
		return
	}

}
