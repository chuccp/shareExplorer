package file

import (
	"log"
	"os"
	"testing"
)

func TestFile(t *testing.T) {

	fs, err := GetRootPath()
	if err == nil {
		for _, v := range fs {
			log.Print(v.Abs())
			ds, err2 := v.ListFile(0)
			if err2 == nil {
				for _, d := range ds {
					println(d.Abs())
					if d.IsDir(){

						fffs,_:=d.ListFile(0)
						for _, fff := range fffs {
							println(fff.Abs())
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

	f,err:=os.Open("D:/")
	if err==nil{
		ds,_:=f.ReadDir(0)
		for _, d := range ds {
			t.Log(d.Name())
		}
	}

}
