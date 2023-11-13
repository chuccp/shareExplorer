package core

import (
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/util"
	"net/http"
)

type Server interface {
	Init(context *Context)
	GetName() string
}
type IRegister interface {
	Range(f func(server Server) bool)
	GetConfig() *util.Config
}

type TraversalServer interface {
	Connect(remoteAddress string) error
	FindRemoteHost(serverName string) (*entity.RemoteHost, error)
	ReverseProxy(remoteHost string, rw http.ResponseWriter, req *http.Request)
	Login(serverName string) error
}
