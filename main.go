package main

import (
	"github.com/chuccp/utils/config"
	"github.com/magiconair/properties"
	"shareExplorer/core"
)

func main()  {

	cfg, err := config.LoadFile("application.properties", properties.UTF8)
	if err==nil{
		shareExplorer:=core.NewShareExplorer(cfg)
		shareExplorer.Start()
	}
}