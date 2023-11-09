package core

import (
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/util"
)

type Server interface {
	Init(context *Context)
	GetName() string
}
type TraversalClient interface {
	Register() error
	Connect() error
	ClientSignIn(username string, password string) error
}
type IRegister interface {
	Range(f func(server Server) bool)
	GetConfig() *util.Config
}

type TraversalServer interface {
	GetUser(username string) *entity.RemoteHost
	GetClient(remoteAddress string) TraversalClient
}
