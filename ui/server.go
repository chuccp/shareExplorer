package ui

import (
	"github.com/chuccp/shareExplorer/core"
)

type Server struct {
	context *core.Context
}

func (s *Server) Init(context *core.Context) {
	webPath := context.GetConfig("ui", "web.path")
	context.StaticHandle("/", webPath)

	///manifest.json
}
func (s *Server) GetName() string {
	return "ui"
}
