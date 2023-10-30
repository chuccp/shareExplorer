package ui

import (
	"github.com/chuccp/shareExplorer/core"
	"log"
)

type Server struct {
	context *core.Context
}

func (s *Server) Init(context *core.Context) {
	webPath := context.GetConfig("ui", "web.path")
	log.Println("webPath:" + webPath)
	context.StaticHandle("/", webPath)

	///manifest.json
}
func (s *Server) GetName() string {
	return "ui"
}
