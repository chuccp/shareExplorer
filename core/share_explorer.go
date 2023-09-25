package core

import (
	"github.com/chuccp/kuic/cert"
	khttp "github.com/chuccp/kuic/http"
	db2 "github.com/chuccp/shareExplorer/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strconv"
)

type ShareExplorer struct {
	engine   *gin.Engine
	register IRegister
	context  *Context
	server   *khttp.Server
}

func CreateShareExplorer(register IRegister) (*ShareExplorer, error) {
	engine := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 允许的域名列表，可以使用 * 来允许所有域名
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
	context := &Context{engine: engine, register: register, server: server, db: db}
	return &ShareExplorer{register: register, engine: engine, context: context, server: server}, nil
}

func (se *ShareExplorer) Start() error {
	se.register.Range(func(server Server) bool {
		server.Init(se.context)
		return true
	})
	keyPath := "keyPath.PEM"
	certPath := "certPath.PEM"
	err := cert.CreateOrReadCert(keyPath, certPath)
	if err != nil {
		return err
	}
	se.engine.Use(func(context *gin.Context) {
		username := context.GetHeader("username")
		if len(username) == 0 {
			context.Next()
		} else {
			traversal, ok := se.context.GetTraversal()
			if ok {
				u := traversal.GetUser(username)
				if u != nil && len(u.RemoteAddr) > 0 {
					proxy, err := se.server.GetReverseProxy(u.RemoteAddr)
					if err == nil {
						proxy.ServeHTTP(context.Writer, context.Request)
					}
				}
			}

		}
	})
	err = se.server.ListenAndServe(certPath, keyPath, se.engine)
	return err
}
