package user

import (
	"errors"
	"github.com/chuccp/kuic/cert"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/db"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/util"
	"github.com/chuccp/shareExplorer/web"
	"gorm.io/gorm"
	"log"
	"net"
)

type admin struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	RePassword  string   `json:"rePassword"`
	IsServer    bool     `json:"isServer"`
	IsNatServer bool     `json:"isNatServer"`
	Addresses   []string `json:"addresses"`
}

type Server struct {
	context *core.Context
}

func (s *Server) GetName() string {
	return "user"
}
func (s *Server) signIn(req *web.Request) (any, error) {
	var admin admin
	req.BodyJson(&admin)
	if len(admin.Username) == 0 {
		return web.ResponseError("用户名不能为空"), nil
	}
	if len(admin.Password) == 0 {
		return web.ResponseError("密码不能为空"), nil
	}

	u, err := s.context.GetDB().GetUserModel().QueryUser(admin.Username, admin.Password)
	if err != nil {
		return nil, err
	}
	if len(u.Username) > 0 {
		sub, err := req.SignedUsername(u.Username)
		if err != nil {
			return nil, err
		}
		return web.ResponseOK(sub), nil
	} else {
		return web.ResponseError("登录失败"), nil
	}

}
func (s *Server) addAdmin(req *web.Request) (any, error) {
	var admin admin
	log.Println("#################", admin)
	req.BodyJson(&admin)
	if len(admin.Username) == 0 {
		return web.ResponseError("用户名不能为空"), nil
	}
	if len(admin.Password) == 0 {
		return web.ResponseError("密码不能为空"), nil
	}
	if len(admin.RePassword) == 0 {
		return web.ResponseError("确认密码不能为空"), nil
	}
	if admin.RePassword != admin.Password {
		return web.ResponseError("两次密码输入不同"), nil
	}
	if admin.IsNatServer || admin.IsServer {
		if admin.Addresses == nil || len(admin.Addresses) == 0 {
			return web.ResponseError("远程节点不能为空"), nil
		}
	}

	err := s.context.GetDB().GetRawDB().Transaction(func(tx *gorm.DB) error {
		err := s.context.GetDB().GetUserModel().NewModel(tx).AddUser(admin.Username, admin.Password, "admin")
		if err != nil {
			return err
		}
		IsServer := "false"
		if admin.IsServer {
			IsServer = "true"
		}
		err = s.context.GetDB().GetConfigModel().NewModel(tx).Create("isServer", IsServer)
		if err != nil {
			return err
		}
		isNatServer := "false"
		if admin.IsNatServer {
			isNatServer = "true"
		}
		err = s.context.GetDB().GetConfigModel().NewModel(tx).Create("isNatServer", isNatServer)
		if err != nil {
			return err
		}
		addressModel := s.context.GetDB().GetAddressModel().NewModel(tx)
		err = addressModel.AddAddress(admin.Addresses)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return web.ResponseError(err.Error()), err
	}
	sub, err := req.SignedUsername(admin.Username)
	if err != nil {
		return nil, err
	}
	err = s.context.GetServerConfig().Init()
	if err != nil {
		return nil, err
	}
	discoverServer, fa := s.context.GetDiscoverServer()
	if fa {
		discoverServer.Start()
	}
	return web.ResponseOK(sub), nil
}

func (s *Server) addClient(req *web.Request) (any, error) {
	var admin admin
	req.BodyJson(&admin)
	if admin.Addresses == nil || len(admin.Addresses) == 0 {
		return web.ResponseError("远程节点不能为空"), nil
	}
	err := s.context.GetDB().GetRawDB().Transaction(func(tx *gorm.DB) error {
		err := s.context.GetDB().GetConfigModel().NewModel(tx).Create("isServer", "false")
		if err != nil {
			return err
		}
		err = s.context.GetDB().GetConfigModel().NewModel(tx).Create("isNatServer", "false")
		if err != nil {
			return err
		}
		addressModel := s.context.GetDB().GetAddressModel().NewModel(tx)
		err = addressModel.AddAddress(admin.Addresses)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	discoverServer, fa := s.context.GetDiscoverServer()
	if fa {
		discoverServer.Start()
	}

	return web.ResponseOK("ok"), nil
}

func (s *Server) info(req *web.Request) (any, error) {
	ServerConfig := s.context.GetServerConfig()
	var system entity.System
	system.HasInit = ServerConfig.HasInit()
	if system.HasInit {
		system.IsServer = ServerConfig.IsServer()
		if system.IsServer {
			username := req.GetTokenUsername()
			if len(username) > 0 {
				system.HasSignIn = true
			} else {
				system.HasSignIn = false
			}
		}
	} else {
		system.RemoteAddress = s.context.GetConfigArray("traversal", "remote.address")
	}
	return &system, nil
}

func (s *Server) reset(req *web.Request) (any, error) {
	err := s.context.GetDB().Reset()
	if err != nil {
		return nil, err
	}
	err = s.context.GetServerConfig().Init()
	if err != nil {
		return nil, err
	}
	ds, b := s.context.GetDiscoverServer()
	if b {
		ds.Stop()
	}
	return web.ResponseOK("ok"), nil
}

func (s *Server) addRemoteAddress(req *web.Request) (any, error) {
	println("addRemoteAddress")
	var addresses []string
	err := req.BodyJson(&addresses)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return web.ResponseError("不能为空"), nil
	}
	addressModel := s.context.GetDB().GetAddressModel()
	err = addressModel.AddAddress(addresses)
	if err != nil {
		return nil, err
	}
	return web.ResponseOK("ok"), nil
}
func (s *Server) connect(req *web.Request) (any, error) {
	address := req.FormValue("address")
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	discoverServer, ok := s.context.GetDiscoverServer()
	if ok {
		err := discoverServer.Connect(addr)
		if err != nil {
			return nil, err
		}
	}
	return web.ResponseOK("ok"), nil
}

func (s *Server) clientSignIn(req *web.Request) (any, error) {

	return web.ResponseOK("ok"), nil
}

func (s *Server) downloadCert(req *web.Request) (any, error) {
	username := req.GetTokenUsername()
	cert, _, err := s.context.GetCertManager().CreateOrReadClientKuicCertFile(username)
	if err != nil {
		return nil, err
	}
	return web.ResponseFile(cert), nil
}
func (s *Server) downloadUserCert(req *web.Request) (any, error) {
	username := req.FormValue("username")
	cert, _, err := s.context.GetCertManager().CreateOrReadClientKuicCertFile(username)
	if err != nil {
		return nil, err
	}
	return web.ResponseFile(cert), nil
}

func (s *Server) uploadUserCert(req *web.Request) (any, error) {
	file, err := req.FormFile("cert")
	if err != nil {
		return nil, err
	}

	data, err := web.ReadAllUploadedFile(file)
	if err != nil {
		return nil, err
	}
	filename := util.MD5(data)
	certPath := "client/" + filename + ".kuic.cert"
	err = web.SaveData(data, certPath)
	if err != nil {
		return nil, err
	}
	c, err := cert.ParseClientKuicCertBytes(data)
	if err != nil {
		return nil, err
	}
	var client entity.Client
	client.Username = c.UserName
	client.ServerName = c.ServerName
	err = s.context.GetDB().GetUserModel().AddClientUser(client.Username, certPath)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	clientCert := s.context.GetClientCert()
	err = clientCert.LoadUser(client.Username)
	if err != nil {
		return nil, err
	}
	return web.ResponseOK(&client), nil
}

func (s *Server) addPath(req *web.Request) (any, error) {
	var path db.Path
	err := req.BodyJson(&path)
	if err != nil {
		return nil, err
	}
	err = s.context.GetDB().GetPathModel().Create(path.Name, path.Path)
	if err != nil {
		return nil, err
	}
	return web.ResponseOK("ok"), nil
}
func (s *Server) editPath(req *web.Request) (any, error) {
	var path db.Path
	err := req.BodyJson(&path)
	if err != nil {
		return nil, err
	}
	err = s.context.GetDB().GetPathModel().Update(int(path.Id), path.Name, path.Path)
	if err != nil {
		return nil, err
	}
	return web.ResponseOK("ok"), nil
}
func (s *Server) deletePath(req *web.Request) (any, error) {
	id := req.FormIntValue("id")
	err := s.context.GetDB().GetPathModel().Delete(uint(id))
	if err != nil {
		return nil, err
	}
	return web.ResponseOK("ok"), nil
}

func (s *Server) queryPath(req *web.Request) (any, error) {
	pageNo := req.FormIntValue("pageNo")
	pageSize := req.FormIntValue("pageSize")
	list, num, err := s.context.GetDB().GetPathModel().QueryPage(pageNo, pageSize)
	if err != nil {
		return nil, err
	}
	pageAble := &web.PageAble{Total: num, List: list}
	return web.ResponseOK(pageAble), nil
}
func (s *Server) queryAllPath(req *web.Request) (any, error) {
	list, num, err := s.context.GetDB().GetPathModel().QueryPage(1, 100)
	if err != nil {
		return nil, err
	}
	pageAble := &web.PageAble{Total: num, List: list}
	return web.ResponseOK(pageAble), nil
}
func (s *Server) queryUser(req *web.Request) (any, error) {
	list, num, err := s.context.GetDB().GetUserModel().QueryPage(1, 100)
	if err != nil {
		return nil, err
	}
	return web.ResponsePage(num, list), nil
}
func (s *Server) queryOnePath(req *web.Request) (any, error) {
	id := req.FormIntValue("id")
	path, err := s.context.GetDB().GetPathModel().QueryById(uint(id))
	return web.ResponseOK(path), err
}

func (s *Server) addUser(req *web.Request) (any, error) {
	var user db.User
	err := req.BodyJson(&user)
	if err != nil {
		return nil, err
	}
	err = s.context.GetDB().GetUserModel().AddGuestUser(user.Username, user.Password, user.PathIds)
	if err != nil {
		return nil, err
	}
	return web.ResponseOK("ok"), err
}

func (s *Server) deleteUser(req *web.Request) (any, error) {
	username := req.FormValue("username")
	err := s.context.GetDB().GetUserModel().DeleteUser(username)
	if err != nil {
		return nil, err
	}
	return web.ResponseOK("ok"), nil
}

func (s *Server) editUser(req *web.Request) (any, error) {
	var user db.User
	err := req.BodyJson(&user)
	if err != nil {
		return nil, err
	}
	err = s.context.GetDB().GetUserModel().EditUser(user.Id, user.Username, user.Password, user.PathIds)
	if err != nil {
		return nil, err
	}
	return web.ResponseOK("ok"), err
}
func (s *Server) queryOneUser(req *web.Request) (any, error) {
	username := req.FormValue("username")
	if len(username) > 0 {
		user, err := s.context.GetDB().GetUserModel().QueryOneUser(username)
		if err != nil {
			return nil, err
		}
		return web.ResponseOK(user), nil
	}
	userId := req.FormIntValue("userId")
	if userId > 0 {
		user, err := s.context.GetDB().GetUserModel().QueryById(uint(userId))
		if err != nil {
			return nil, err
		}
		return web.ResponseOK(user), nil
	}
	return nil, errors.New("参数有错")
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	context.Get("/user/info", s.info)
	context.Get("/user/reset", s.reset)
	context.Post("/user/addAdmin", s.addAdmin)
	context.Post("/user/addClient", s.addClient)

	context.Post("/user/clientSignIn", s.clientSignIn)
	context.Post("/user/addRemoteAddress", s.addRemoteAddress)
	context.Get("/user/connect", s.connect)
	context.Get("/user/downloadCert", s.downloadCert)
	context.Get("/user/downloadUserCert", s.downloadUserCert)
	context.Post("/user/uploadUserCert", s.uploadUserCert)

	context.Get("/user/queryUser", s.queryUser)
	context.Post("/user/addUser", s.addUser)
	context.Get("/user/deleteUser", s.deleteUser)
	context.Post("/user/editUser", s.editUser)
	context.Get("/user/queryOneUser", s.queryOneUser)

	context.Get("/user/queryOnePath", s.queryOnePath)
	context.Post("/user/addPath", s.addPath)
	context.Post("/user/editPath", s.editPath)
	context.Get("/user/deletePath", s.deletePath)

	context.GetRemote("/user/queryPath", s.queryPath)
	context.GetRemote("/user/queryAllPath", s.queryAllPath)
	context.PostRemote("/user/signIn", s.signIn)
}
