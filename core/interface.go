package core

import (
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/util"
	"net"
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
	Ping(address *net.UDPAddr) error
	Stop()
	FindStatus(servername string, isStart bool) *entity.NodeStatus
}
