package discover

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
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
	context    *core.Context
	table      *Table
	nodeSearch *nodeSearch
	call       *call
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
func (s *Server) findValue(req *web.Request) (any, error) {
	var findValue FindValue
	err := req.BodyJson(&findValue)
	if err != nil {
		return nil, err
	}
	ns := s.table.FindValue(findValue.Target, findValue.Distances)
	return web.ResponseOK(wrapResponseNodes(ns)), nil
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	s.call = &call{httpClient: core.NewHttpClient(context)}
	if !s.context.GetServerConfig().HasInit() {
		return
	}
	s.context.SetDiscoverServer(s)
	s.Start()
}
func (s *Server) nodeList(req *web.Request) (any, error) {
	nodeType := req.FormIntValue("nodeType")
	pageNo := req.FormIntValue("pageNo")
	pageSize := req.FormIntValue("pageSize")
	list, num := s.table.nodePage(nodeType, pageNo, pageSize)
	return web.ResponsePage(int64(num), wrapExNodes(list)), nil
}
func (s *Server) connect(req *web.Request) (any, error) {
	return web.ResponseOK("ok"), nil
}

func (s *Server) FindStatus() *entity.NodeStatus {
	return s.nodeSearch.nodeStatus
}

func (s *Server) Connect(address *net.UDPAddr) error {
	_, err := s.call.httpClient.GetRequest(address, "/discover/connect")
	if err != nil {
		return err
	}
	return nil
}
func (s *Server) nodeStatus(req *web.Request) (any, error) {
	var nodeStatus NodeStatus
	err := req.BodyJson(&nodeStatus)
	if err != nil {
		return nil, err
	}
	if s.context.GetServerConfig().IsClient() {

	}
	return web.ResponseOK("ok"), nil
}
func (s *Server) Start() {
	if s.context.GetServerConfig().IsNatServer() {
		s.context.Post("/discover/register", s.register)
		s.context.Post("/discover/connect", s.connect)
		s.context.Get("/discover/nodeList", s.nodeList)
		s.context.Post("/discover/findNode", s.findNode)
		s.context.Post("/discover/findValue", s.findValue)
	}
	s.context.Get("/discover/nodeStatus", s.nodeStatus)
	servername := s.context.GetCertManager().GetServerName()
	localNode, err := createLocalNode(servername)
	if err != nil {
		log.Println(err)
		return
	}
	s.table = NewTable(s.context, localNode, s.call)
	s.table.run()
	s.nodeSearch = newNodeSearch(s.table, localNode)
	if s.context.GetServerConfig().IsClient() {
		s.nodeSearch.run()
	}
}
func (s *Server) Stop() {
	if s.table != nil {
		s.table.stop()
	}
	if s.nodeSearch != nil {
		s.nodeSearch.stop()
	}
}

func (s *Server) GetName() string {
	return "discover"
}
