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
	var register Register
	err := req.BodyJson(&register)
	if err != nil {
		return nil, err
	}
	node, err := wrapNodeFRegister(&register, req.GetRemoteAddress())
	if err != nil {
		return nil, err
	}
	s.tableGroup.addSeenNode(wrapNode(node))
	n := s.tableGroup.GetOneTable().self()
	return web.ResponseOK(wrapResponseNode(n)), nil
}
func (s *Server) findNode(req *web.Request) (any, error) {
	var findNode FindNode
	err := req.BodyJson(&findNode)
	if err != nil {
		return nil, err
	}
	addr, err := net.ResolveUDPAddr("udp", req.GetRemoteAddress())
	if err != nil {
		return nil, err
	}
	ns := s.tableGroup.GetOneTable().HandleFindNode(addr.IP, &findNode)
	return web.ResponseOK(wrapResponseNodes(ns)), nil
}
func (s *Server) queryNode(req *web.Request) (any, error) {
	return web.ResponseOK(""), nil
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	s.context.Post("/discover/register", s.register)
	s.context.Post("/discover/findNode", s.findNode)
	s.context.Post("/discover/queryNode", s.queryNode)
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
