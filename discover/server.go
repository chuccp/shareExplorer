package discover

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/web"
	"log"
	"net"
)

type Server struct {
	context    *core.Context
	tableGroup *TableGroup
}

func (s *Server) register(req *web.Request) (any, error) {
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
	s.context.Post("/discover/register", s.register)
	s.tableGroup = NewTableGroup(context)
	table := s.tableGroup.AddTable(newLocalNode("111111"))
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2156")
	if err == nil {
		table.addNursery(addr)
	} else {
		log.Println(err)
	}
	s.tableGroup.run()
}

func (s *Server) GetName() string {
	return "discover"
}
