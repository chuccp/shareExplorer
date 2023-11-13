package traversal

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/util"
	"github.com/chuccp/shareExplorer/web"
	"net/http"
)

type Server struct {
	store         *ClientStore
	context       *core.Context
	clientManager *ClientManager
	request       *Request
}

func (server *Server) register(req *web.Request) (any, error) {
	var user entity.RemoteHost
	err := req.BodyJson(&user)
	if err != nil {
		return nil, err
	}
	user.RemoteAddr = req.GetRemoteAddress()
	user.UpdateTime = util.NowTime()
	user.CreateTime = util.NowTime()
	user.LastLiveTime = util.NowTime()
	user.IsOnline = true
	server.store.AddUser(&user)
	return "ok", nil
}
func (server *Server) queryRemoteHostList(req *web.Request) (any, error) {
	page := req.GetPage()
	return server.store.QueryPage(page), nil
}
func (server *Server) queryRemoteHostOne(req *web.Request) (any, error) {
	serverName := req.FormValue("serverName")
	u := server.GetRemoteHost(serverName)
	if u != nil {
		return u, nil
	}
	return nil, web.NotFound
}

func (server *Server) GetRemoteHost(serverName string) *entity.RemoteHost {
	u, ok := server.store.Query(serverName)
	if ok {
		return u
	}
	return nil
}

func (server *Server) Connect(remoteAddress string) error {
	_, err := server.request.RequestString(remoteAddress, "/traversal/connect")
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) FindRemoteHost(serverName string) (*entity.RemoteHost, error) {

	server.clientManager.FindRemoteHost(serverName)

	return nil, nil
}

func (server *Server) ReverseProxy(remoteHost string, rw http.ResponseWriter, req *http.Request) {
	req.Header.Del("Origin")
	req.Header.Del("Referer")
	proxy, err := server.context.GetReverseProxy(remoteHost, nil)
	if err != nil {
		return
	}
	proxy.ServeHTTP(rw, req)
}
func (server *Server) Login(serverName string) error {

	return nil
}

func (server *Server) connect(req *web.Request) (any, error) {
	return web.ResponseOK("ok"), nil
}

func (server *Server) runClient() {
	go server.clientManager.Run()
}

func (server *Server) Init(context *core.Context) {
	server.context = context
	server.store = newClientStore()
	server.request = NewRequest(server.context)
	server.clientManager = NewClientManager(context)
	server.context.SetTraversal(server)
	serverConfig := server.context.GetServerConfig()
	if serverConfig.IsServer() && serverConfig.IsNatServer() && serverConfig.IsNatClient() {
		remoteHost := entity.NewRemoteHost(context.GetCertManager().GetServerName(), "0.0.0.0:0")
		server.store.AddUser(remoteHost)
	}
	if serverConfig.IsNatServer() {
		context.Post("/traversal/register", server.register)
		context.Get("/traversal/connect", server.connect)
		context.Get("/traversal/queryRemoteHostList", server.queryRemoteHostList)
		context.Get("/traversal/queryRemoteHostOne", server.queryRemoteHostOne)
	}
	if serverConfig.IsNatClient() {
		server.runClient()
	}
}
func (server *Server) GetName() string {
	return "traversal"
}
