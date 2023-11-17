package discover

import "github.com/chuccp/shareExplorer/core"

type Server struct {
	context    *core.Context
	tableGroup *TableGroup
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	s.tableGroup = NewTableGroup(context)
	s.tableGroup.run()
}

func (s *Server) GetName() string {
	return "discover"
}
