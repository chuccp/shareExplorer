package discover

import (
	"github.com/chuccp/kuic/cert"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/web"
	"go.uber.org/zap"
	"net"
)

const (
	client = iota + 1
	server
	natServer
)

type Server struct {
	context          *core.Context
	table            *Table
	nodeSearchManage *nodeSearchManage
	call             *call
	localNode        *Node
	servername       string
}

func (s *Server) register(req *web.Request) (any, error) {
	var register Register
	v, err := req.BodyJson(&register)
	s.context.GetLog().Debug("register", zap.ByteString("Request", v), zap.Error(err))
	if err != nil {
		s.context.GetLog().Error("register", zap.Error(err), zap.ByteString("body", v))
		return nil, err
	}
	node, err := wrapNodeFRegister(&register, req.GetRemoteAddress())
	if err != nil {
		s.context.GetLog().Error("register", zap.Error(err))
		return nil, err
	}
	s.table.AddNode(node)
	n := s.table.self()
	return web.ResponseOK(wrapResponseNode(n)), nil
}
func (s *Server) findNode(req *web.Request) (any, error) {
	var findNode FindNode
	v, err := req.BodyJson(&findNode)
	if err != nil {
		s.context.GetLog().Error("findNode", zap.Error(err), zap.ByteString("body", v))
		return nil, err
	}
	addr, err := net.ResolveUDPAddr("udp", req.GetRemoteAddress())
	if err != nil {
		s.context.GetLog().Error("findNode", zap.Error(err))
		return nil, err
	}
	findNode.addr = addr
	ns := s.table.FindNode(&findNode)
	return web.ResponseOK(wrapResponseNodes(ns)), nil
}
func (s *Server) findServer(req *web.Request) (any, error) {
	var findServer FindServer
	v, err := req.BodyJson(&findServer)
	if err != nil {
		s.context.GetLog().Error("findServer", zap.Error(err), zap.ByteString("body", v))
		return nil, err
	}
	id, _ := wrapIdFName(findServer.Target)
	n, ns := s.table.FindServer(id, findServer.Distances)
	return web.ResponseOK(wrapFindServerResponse(n, ns)), nil
}

func (s *Server) nodeStatus(req *web.Request) (any, error) {
	var nodeStatus NodeStatus
	nodeStatus.Id = s.localNode.ServerName()
	serverConfig := s.context.GetServerConfig()
	if serverConfig.IsServer() {
		nodeStatus.IsServer = "true"
	} else {
		nodeStatus.IsServer = "false"
	}
	if serverConfig.IsNatServer() {
		nodeStatus.IsNatServer = "true"
	} else {
		nodeStatus.IsNatServer = "false"
	}
	if serverConfig.IsClient() {
		nodeStatus.IsClient = "true"
	} else {
		nodeStatus.IsClient = "false"
	}
	return web.ResponseOK(&nodeStatus), nil
}

func (s *Server) findUserServer(req *web.Request) (any, error) {

	username := req.FormValue("username")
	code := req.FormValue("code")

	s.context.GetLog().Debug("findUserServer", zap.String("code", code), zap.String("username", username))

	discoverServer, fa := s.context.GetDiscoverServer()
	if fa {
		user, err := s.context.GetDB().GetUserModel().QueryOneUser(username, code)
		if err != nil {
			return nil, err
		}
		_, ce, err := cert.ParseClientKuicCertFile(user.CertPath)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		status, err := discoverServer.FindStatusWait(ce.ServerName, true)
		if err != nil {
			return err, nil
		}
		return web.ResponseOK(status), nil
	}

	return web.ResponseOK("ok"), nil
}
func (s *Server) Servername() string {
	return s.servername
}
func (s *Server) Init(context *core.Context) {
	s.context = context
	s.servername = s.context.GetCertManager().GetServerName()
	s.Start()
	s.context.Post("/discover/register", s.register)
	s.context.Post("/discover/connect", s.connect)
	s.context.Post("/discover/nodeStatus", s.nodeStatus)
	s.context.GetRemote("/discover/nodeNatServerList", s.nodeNatServerList)
	s.context.GetRemote("/discover/nodeServerList", s.nodeServerList)
	s.context.Post("/discover/findNode", s.findNode)
	s.context.Post("/discover/findServer", s.findServer)
	s.context.Post("/discover/findUserServer", s.findUserServer)
}
func (s *Server) nodeNatServerList(req *web.Request) (any, error) {
	pageNo := req.FormIntValue("pageNo")
	pageSize := req.FormIntValue("pageSize")
	list, num := s.table.queryNatServerForPage(pageNo, pageSize)
	return web.ResponsePage(int64(num), NodeToExNodes(list)), nil
}

func (s *Server) nodeServerList(req *web.Request) (any, error) {
	pageNo := req.FormIntValue("pageNo")
	pageSize := req.FormIntValue("pageSize")
	list, num := s.table.queryServerForPage(pageNo, pageSize)
	return web.ResponsePage(int64(num), NodeToExNodes(list)), nil
}

func (s *Server) connect(req *web.Request) (any, error) {
	return web.ResponseOK("ok"), nil
}

func (s *Server) FindStatus(servername string, isStart bool) *entity.NodeStatus {
	id, _ := wrapIdFName(servername)
	return s.nodeSearchManage.FindNodeStatus(id, isStart)
}
func (s *Server) FindStatusWait(servername string, isWait bool) (*entity.NodeStatus, error) {
	s.context.GetLog().Debug("FindStatusWait", zap.String("servername", servername))
	id, err := StringToId(servername)
	if err != nil {
		return nil, err
	}
	return s.nodeSearchManage.FindWaitNodeStatus(id, isWait), nil
}

func (s *Server) QueryStatus(servername ...string) []*entity.NodeStatus {
	return s.nodeSearchManage.QueryStatus(servername...)
}

func (s *Server) Ping(address *net.UDPAddr) error {
	err := s.call.ping(address)
	if err != nil {
		s.context.GetLog().Error("Ping", zap.Error(err))
		return err
	}
	return nil
}
func (s *Server) ReStart() {
	s.table.stop()
	if s.nodeSearchManage != nil {
		s.nodeSearchManage.stopAll()
	}
	s.Start()
}
func (s *Server) Start() {
	id, err := StringToId(s.servername)
	if err != nil {
		s.context.GetLog().Error("Init", zap.Error(err))
		return
	}
	s.localNode = NewLocalNode(id, s.context.GetServerConfig())
	s.call = newCall(s.localNode, core.NewHttpClient(s.context), s.context)
	s.context.SetDiscoverServer(s)
	s.table = NewTable(s.context, s.localNode, s.call)
	s.nodeSearchManage = NewNodeSearchManage(s.context, s.table)
	if !s.context.GetServerConfig().HasInit() {
		return
	}
	go s.table.run()
	go s.nodeSearchManage.run()
}
func (s *Server) Stop() {
	if s.table != nil {
		s.table.stop()
	}
	if s.nodeSearchManage != nil {
		s.nodeSearchManage.stopAll()
	}
}

func (s *Server) GetName() string {
	return "discover"
}
