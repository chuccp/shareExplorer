package traversal

import (
	"github.com/chuccp/shareExplorer/core"
	user2 "github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/util"
	"github.com/chuccp/shareExplorer/web"
	"log"
)

type Server struct {
	store   *Store
	context *core.Context
}

func (s *Server) GetClient(remoteAddress string) core.TraversalClient {
	return newClient(s.context, remoteAddress)
}

func (s *Server) input(req *web.Request) (any, error) {
	var user user2.RemoteHost
	err := req.BodyJson(&user)
	if err != nil {
		return nil, err
	}
	user.RemoteAddr = req.GetRemoteAddress()
	user.UpdateTime = util.NowTime()
	user.CreateTime = util.NowTime()
	user.LastLiveTime = util.NowTime()
	user.IsOnline = true
	s.store.AddUser(&user)
	return "ok", nil
}
func (s *Server) queryList(req *web.Request) (any, error) {
	page := req.GetPage()
	return s.store.QueryPage(page), nil
}
func (s *Server) queryOne(req *web.Request) (any, error) {
	username := req.FormValue("username")
	u := s.GetUser(username)
	if u != nil {
		return u, nil
	}
	return nil, web.NotFound
}

func (s *Server) GetUser(username string) *user2.RemoteHost {
	u, ok := s.store.Query(username)
	if ok {
		return u
	}
	return nil
}
func (s *Server) connect(req *web.Request) (any, error) {

	log.Println("检测=================来了")

	return web.ResponseOK("ok"), nil
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	s.store = newStore()
	context.SetTraversal(s)
	context.Post("/traversal/register", s.input)
	context.Get("/traversal/connect", s.connect)
	context.Get("/traversal/queryList", s.queryList)
	context.Get("/traversal/queryOne", s.queryOne)
	//go s.client.start()
}
func (s *Server) GetName() string {
	return "traversal"
}
