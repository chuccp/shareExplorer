package core

import (
	"github.com/chuccp/kuic/cert"
	khttp "github.com/chuccp/kuic/http"
	db2 "github.com/chuccp/shareExplorer/db"
	"github.com/chuccp/shareExplorer/web"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

type ShareExplorer struct {
	engine       *gin.Engine
	register     IRegister
	context      *Context
	server       *khttp.Server
	certManager  *cert.Manager
	serverConfig *ServerConfig
	clientCert   *ClientCert
}

func CreateShareExplorer(register IRegister) (*ShareExplorer, error) {
	engine := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 允许的域名列表，可以使用 * 来允许所有域名
	config.AllowHeaders = []string{"*"} // 允
	engine.Use(cors.New(config))
	port, err := register.GetConfig().GetInt("core", "local.port")
	if err != nil {
		return nil, err
	}
	server, err := khttp.CreateServer("0.0.0.0:" + strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	fileName := register.GetConfig().GetStringOrDefault("db", "file.name", "share_explorer.db")
	db, err := db2.CreateDb(fileName)
	if err != nil {
		return nil, err
	}
	certManager := cert.NewManager("cert")
	serverConfig := NewServerConfig(db.GetConfigModel())

	logger, err := initLogger("log/run.log")
	if err != nil {
		return nil, err
	}
	context := &Context{log: logger, serverConfig: serverConfig, engine: engine, register: register, server: server, db: db, paths: make(map[string]any), remotePaths: make(map[string]any), certManager: certManager}
	clientCert := NewClientCert(context, db.GetUserModel())
	context.clientCert = clientCert
	context.digestAuth = web.NewDigestAuthenticator("share_explorer", context.Secret)
	return &ShareExplorer{clientCert: clientCert, register: register, engine: engine, context: context, server: server, certManager: certManager, serverConfig: serverConfig}, nil
}

func (se *ShareExplorer) Start() error {
	//证书初始化
	err := se.certManager.Init()
	if err != nil {
		return err
	}
	//服务配置初始化
	_, err = se.serverConfig.Init()
	if err != nil {
		return err
	}
	se.context.GetLog().Debug("IsClient", zap.Bool("IsClient", se.serverConfig.IsClient()))
	//加载客户端证书
	if se.serverConfig.IsClient() {
		se.clientCert.init()
	}
	se.context.RemoteHandle()
	se.register.Range(func(server Server) bool {
		server.Init(se.context)
		return true
	})
	if err != nil {
		return err
	}

	err = se.server.ListenAndServeWithKuicTls(se.certManager, se.engine)
	return err
}
