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
	id, err := wrapIdFName(node.ServerName)
	if err != nil {
		return nil, err
	}
	node.SetID(id)
	s.tableGroup.addSeenNode(wrapNode(&node))
	n := s.tableGroup.GetOneTable().localNode
	return web.ResponseOK(n), nil
}
func (s *Server) Init(context *core.Context) {
	s.context = context
	s.context.Post("/discover/register", s.register)
	s.tableGroup = NewTableGroup(context)
	localNode, err := createLocalNode("123456789abc")
	if err != nil {
		log.Println(err)
		return
	}
	table := s.tableGroup.AddTable(localNode)
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
