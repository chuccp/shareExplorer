package core

import (
	"github.com/chuccp/shareExplorer/util"
)

type Server interface {
	Init(context *Context)
	GetName() string
}
type IRegister interface {
	Range(f func(server Server) bool)
	GetConfig() *util.Config
}

type DiscoverServer interface {
	Start()
	Connect(address string) error
	Stop()
	FindAddress() (string, error)
}
