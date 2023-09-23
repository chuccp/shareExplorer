package user

import (
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/web"
	"gorm.io/gorm"
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
	system.RemoteAddress = s.context.GetConfigArray("traversal", "remote.address")
	return &system, nil
}

func (s *Server) Init(context *core.Context) {
	s.context = context
	context.Get("/user/info", s.info)
	context.Post("/user/addAdmin", s.addAdmin)
}
