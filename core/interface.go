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
	ReStart()
	Ping(address *net.UDPAddr) error
	Stop()
	Servername() string
	FindStatusWait(servername string, isWait bool) (*entity.NodeStatus, error)

	QueryStatus(servername ...string) []*entity.NodeStatus
}
