package main

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/file"
	"github.com/chuccp/shareExplorer/traversal"
	"github.com/chuccp/shareExplorer/user"
)

func main() {

	cfg, err := core.LoadFile("config.ini")
	if err != nil {
		return
	}
	var register = core.NewRegister(cfg)
	register.AddServer(&file.Server{})
	register.AddServer(&user.Server{})
	register.AddServer(&traversal.Server{})
	shareExplorer, err := register.Create()
	if err != nil {
		panic(err)
		return
	}
	shareExplorer.Start()

}
