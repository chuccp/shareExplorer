package file

import (
	"log"
	"path/filepath"
	"testing"
)

func TestGetRelativePath(t *testing.T)  {

	v,_:=filepath.Rel("D:\\aaa","D:\\aaa\\bb\\111.json")

	t.Log(v)
}

func TestFile(t *testing.T) {

	fs, err := GetRootPath()
	if err == nil {
		for _, v := range fs {
			log.Print(v.Abs())
			ds, err2 := v.ListAllFile()
			if err2 == nil {
				for _, d := range ds {
					println(d.Abs())
					if d.IsDir(){
						faffs,_:=d.ListAllFile()
						for _, fff := range faffs {
							println(fff.Name())
						}

					}
				}
			}else{
				log.Print(err2)
			}
		}
	}
}


func TestFile2(t *testing.T) {

	f,err:=NewFile("C:/")
	if err==nil{
		log.Println(f.Abs())
		ds,err:=f.ListAllFile()
		log.Println(err)
		for _, d := range ds {
			t.Log(d.Name())
		}
	}

}
