package io

import (
	"log"
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
