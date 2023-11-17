package discover

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/web"
)

type Server struct {
	context    *core.Context
	tableGroup *TableGroup
}

func (s *Server) ping(req *web.Request) (any, error) {
	var node Node
	err := req.BodyJson(&node)
	if err != nil {
		return nil, err
	}
	s.tableGroup.addSeenNode(wrapNode(&node))

	return web.ResponseOK("ok"), nil
}
func (s *Server) Init(context *core.Context) {
	s.context = context
	s.context.Post("/discover/ping", s.ping)
	s.tableGroup = NewTableGroup(context)
	s.tableGroup.run()
}

func (s *Server) GetName() string {
	return "discover"
}
