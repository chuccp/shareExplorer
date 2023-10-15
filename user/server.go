package user

import (
	"github.com/chuccp/kuic/util"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/web"
	"gorm.io/gorm"
	"net"
)

type admin struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	RePassword  string `json:"rePassword"`
	IsNatClient bool   `json:"isNatClient"`
	IsNatServer bool   `json:"isNatServer"`
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

	err := s.context.GetDB().GetRawDB().Transaction(func(tx *gorm.DB) error {
		err := s.context.GetDB().GetUserModel().NewModel(tx).AddUser(admin.Username, admin.Password, "admin")
		if err != nil {
			return err
		}
		isNatClient := "false"
		if admin.IsNatClient {
			isNatClient = "true"
		}
		err = s.context.GetDB().GetConfigModel().NewModel(tx).Create("isNatClient", isNatClient)
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
		return nil
	})
	if err != nil {
		return web.ResponseError(err.Error()), err
	}
	return web.ResponseOK("添加成功"), nil
}

func (s *Server) info(req *web.Request) (any, error) {
	exist := s.context.GetDB().GetUserModel().IsExist()
	var system entity.System
	system.HasInit = exist
	if exist {
		username := req.GetTokenUsername()
		if len(username) > 0 {
			system.HasSignIn = true
		} else {
			system.HasSignIn = false
		}
	}
	system.RemoteAddress = s.context.GetConfigArray("traversal", "remote.address")
	return &system, nil
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
	traversalServer, ok := s.context.GetTraversal()
	if ok {
		traversalClient := traversalServer.GetClient(addr.String())
		err := traversalClient.Connect()
		if err != nil {
			return nil, err
		}
	}
	return web.ResponseOK("ok"), nil
}
func (s *Server) downloadCert(req *web.Request) (any, error) {

	cert, err := core.InitClientCert("share")
	if err != nil {
		return nil, err
	}
	pem, err := core.GenerateClientGroupPem(s.context.GetCert().CaPath, cert.CertPath, cert.KeyPath)
	if err != nil {
		return nil, err
	}
	err = util.WriteFile("share.group.key", pem)
	if err != nil {
		return nil, err
	}
	return web.ResponseFile("share.group.key"), nil
}
func (s *Server) addPath(req *web.Request) (any, error) {

	return nil, nil
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	context.Get("/user/info", s.info)
	context.Post("/user/addAdmin", s.addAdmin)
	context.Post("/user/signIn", s.signIn)
	context.Post("/user/addRemoteAddress", s.addRemoteAddress)
	context.Get("/user/connect", s.connect)
	context.Get("/user/downloadCert", s.downloadCert)
	context.Post("/user/addPath", s.addPath)
}
