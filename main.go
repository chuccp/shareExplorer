package main

import (
	log "github.com/chuccp/coke-log"
	"github.com/chuccp/cokePush/config"
	"github.com/magiconair/properties"
	"shareExplorer/core"
)

func main()  {

	cfg, err := config.LoadFile("application.properties", properties.UTF8)
	if err==nil{
		shareExplorer:=core.NewShareExplorer(cfg)
		err:=shareExplorer.Start()
		if err!=nil{
			log.Error("启动失败",err)
		}
	}
}