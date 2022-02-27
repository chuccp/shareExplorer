package file

import (
	log "github.com/chuccp/coke-log"
	"testing"
)

func TestExists(t *testing.T) {
	file,err:=NewFile("static")
	if err==nil{
		log.Info(file.Name(),"====",file.IsDir())
	}
}

func TestOpenOrCreateFile(t *testing.T) {
	file,err:=NewFile("static/sss.ini")
	log.Info("!!!========",err)
	if err==nil{
		data,flag,err3:=file.Read()
		if flag && err3==nil{
			log.Info(data)
		}else{
			log.Info(flag,err3)
		}
	}else{
		log.Info(err)
	}
}

func TestGetRootPath(t *testing.T) {
	files,err:=GetRootPath()
	if err==nil{
		for _, file := range files {
			t.Logf(file.Name())
		}
	}
}
func TestFile_Read(t *testing.T) {


}