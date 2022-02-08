package brower

import (
	"log"
	"os"
	"testing"
)

func TestFile(t *testing.T) {


	f,err:=os.Open("c:/")
	if err==nil{
		fStat,_:=f.Stat()
		log.Print(fStat.IsDir())
		dirs,err:=f.ReadDir(0)
		if err==nil{
			for _,v:=range dirs{

				log.Print(v.Name())

			}
		}


	}else{
		log.Print(err)
	}

}