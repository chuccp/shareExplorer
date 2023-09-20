package user

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/web"
)

type Server struct {
}

func (s *Server) GetName() string {
	return "user"
}

func (s *Server) info(req *web.Request) (any, error) {
	var system entity.System
	system.HasConfig = false
	return &system, nil
}

func (s *Server) Init(context *core.Context) {
	context.Get("/user/info", s.info)
}
