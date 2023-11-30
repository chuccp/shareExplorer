package discover

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/web"
	"log"
	"net"
)

const (
	client = iota + 1
	natClient
	natServer
)

type Server struct {
	context *core.Context
	table   *Table
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
	s.table.addSeenNode(wrapNode(node))
	n := s.table.self()
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
	ns := s.table.HandleFindNode(addr.IP, &findNode)
	return web.ResponseOK(wrapResponseNodes(ns)), nil
}
func (s *Server) queryNode(req *web.Request) (any, error) {
	return web.ResponseOK(""), nil
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	if !s.context.GetServerConfig().HasInit() {
		return
	}
	s.run()
}
func (s *Server) nodeList(req *web.Request) (any, error) {
	nodeType := req.FormIntValue("nodeType")
	log.Println(nodeType)
	return web.ResponseOK(""), nil
}

func (s *Server) run() {
	s.context.Post("/discover/register", s.register)
	s.context.Post("/discover/nodeList", s.nodeList)
	s.context.Post("/discover/findNode", s.findNode)
	s.context.Post("/discover/queryNode", s.queryNode)

	servername := s.context.GetCertManager().GetServerName()
	localNode, err := createLocalNode(servername)
	if err != nil {
		log.Println(err)
		return
	}
	s.table = NewTable(s.context, localNode)
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2156")
	if err == nil {
		s.table.addNursery(addr)
	} else {
		log.Println(err)
	}
	s.table.run()
}

func (s *Server) GetName() string {
	return "discover"
}
