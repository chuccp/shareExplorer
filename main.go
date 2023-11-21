package main

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/discover"
	"github.com/chuccp/shareExplorer/file"
	"github.com/chuccp/shareExplorer/ui"
	"github.com/chuccp/shareExplorer/user"
	"github.com/chuccp/shareExplorer/util"
)

func main() {

	cfg, err := util.LoadFile("config.ini")
	if err != nil {
		return
	}
	var register = core.NewRegister(cfg)
	register.AddServer(&ui.Server{})
	register.AddServer(&file.Server{})
	register.AddServer(&user.Server{})
	register.AddServer(&discover.Server{})
	shareExplorer, err := register.Create()
	if err != nil {
		panic(err)
		return
	}
	err = shareExplorer.Start()
	if err != nil {
		panic(err)
		return
	}
}
