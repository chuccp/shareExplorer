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
	ReStart()
	Ping(address *net.UDPAddr) error
	Stop()
	Servername() string
	FindStatus(servername string, isStart bool) *entity.NodeStatus
	FindStatusWait(servername string) (*entity.NodeStatus, error)

	QueryStatus(servername ...string) []*entity.NodeStatus
}
