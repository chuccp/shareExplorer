package core

import (
	"fmt"
	"github.com/chuccp/kuic/cert"
	khttp "github.com/chuccp/kuic/http"
	"github.com/chuccp/shareExplorer/db"
	"github.com/chuccp/shareExplorer/util"
	"github.com/chuccp/shareExplorer/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"net"
	"net/http"
	"path"
	"runtime/debug"
	"strings"
	"time"
)

const FindWaitTime = 20 * time.Second

type Context struct {
	engine         *gin.Engine
	register       IRegister
	server         *khttp.Server
	discoverServer DiscoverServer
	db             *db.DB
	paths          map[string]any
	remotePaths    map[string]any
	certManager    *cert.Manager
	serverConfig   *ServerConfig
	digestAuth     *web.DigestAuth
	clientCert     *ClientCert
	log            *zap.Logger
}

func (c *Context) GetConfig(section, name string) string {
	return c.register.GetConfig().GetString(section, name)
}
func (c *Context) GetConfigArray(section, name string) []string {
	values := c.register.GetConfig().GetString(section, name)
	vs := strings.Split(values, ",")
	return util.RemoveRepeatElement(vs)
}
func (c *Context) GetDB() *db.DB {
	return c.db
}
func (c *Context) GetLog() *zap.Logger {
	return c.log
}
func (c *Context) GetServerConfig() *ServerConfig {
	return c.serverConfig
}
func (c *Context) GetCertManager() *cert.Manager {
	return c.certManager
}
func (c *Context) GetClientCert() *ClientCert {
	return c.clientCert
}
func (c *Context) GetDigestAuth() *web.DigestAuth {
	return c.digestAuth
}
func (c *Context) JustCheck(relativePath string, handler web.HandlerFunc) web.HandlerFunc {
	return c.digestAuth.JustCheck(relativePath, handler)
}
func (c *Context) CheckAuth(relativePath string, handler web.HandlerFunc) web.HandlerFunc {
	return c.digestAuth.CheckAuth(relativePath, handler)
}

func (c *Context) SetDiscoverServer(discoverServer DiscoverServer) {
	c.discoverServer = discoverServer
}
func (c *Context) GetDiscoverServer() (DiscoverServer, bool) {
	return c.discoverServer, c.discoverServer != nil
}
func (c *Context) GetConfigInt(section, name string) (int, error) {
	return c.register.GetConfig().GetInt(section, name)
}

func (c *Context) GetConfigIntOrDefault(section, name string, defaultValue int) int {
	return c.register.GetConfig().GetIntOrDefault(section, name, defaultValue)
}
func (c *Context) GetConfigInt64OrDefault(section, name string, defaultValue int64) int64 {
	return c.register.GetConfig().GetInt64OrDefault(section, name, defaultValue)
}

func (c *Context) GetHttpClient(address *net.UDPAddr) (*khttp.Client, error) {
	return c.server.GetHttpClient(address)
}
func (c *Context) justChecks(relativePath string, handlers ...web.HandlerFunc) []web.HandlerFunc {
	var hs = make([]web.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hs[i] = c.JustCheck(relativePath, handler)
	}
	return hs
}

func (c *Context) checkAuths(relativePath string, handlers ...web.HandlerFunc) []web.HandlerFunc {
	var hs = make([]web.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hs[i] = c.CheckAuth(relativePath, handler)
	}
	return hs
}

func (c *Context) Post(relativePath string, handlers ...web.HandlerFunc) {
	_, ok := c.paths[relativePath]
	if ok {
		return
	}
	c.paths[relativePath] = true
	c.engine.POST(relativePath, web.ToGinHandlerFuncs(handlers)...)
}
func (c *Context) PostAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.Post(relativePath, c.justChecks(relativePath, handlers...)...)
}
func (c *Context) Get(relativePath string, handlers ...web.HandlerFunc) {
	_, ok := c.paths[relativePath]
	if ok {
		return
	}
	c.paths[relativePath] = true
	c.engine.GET(relativePath, web.ToGinHandlerFuncs(handlers)...)
}

func (c *Context) GetAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.Get(relativePath, c.justChecks(relativePath, handlers...)...)
}

func (c *Context) GetCheckAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.Get(relativePath, c.checkAuths(relativePath, handlers...)...)
}

func (c *Context) Any(relativePath string, handlers ...web.HandlerFunc) {
	_, ok := c.paths[relativePath]
	if ok {
		return
	}
	c.paths[relativePath] = true
	c.engine.Any(relativePath, web.ToGinHandlerFuncs(handlers)...)
	httpMethods := []string{"PROPFIND", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "PROPPATCH"}
	for _, method := range httpMethods {
		c.engine.Handle(method, relativePath, web.ToGinHandlerFuncs(handlers)...)
	}
}
func (c *Context) AnyAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.Any(relativePath, c.justChecks(relativePath, handlers...)...)
}
func (c *Context) AnyRemote(relativePath string, handlers ...web.HandlerFunc) {
	c.Any(relativePath, handlers...)
	c.remotePaths[relativePath] = true
	c.paths[relativePath] = true
}
func (c *Context) AnyRemoteAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.AnyAuth(relativePath, handlers...)
	c.remotePaths[relativePath] = true
	c.paths[relativePath] = true
}
func (c *Context) GetRemote(relativePath string, handlers ...web.HandlerFunc) {
	c.Get(relativePath, handlers...)
	c.remotePaths[relativePath] = true
	c.paths[relativePath] = true
}
func (c *Context) GetRemoteAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.GetAuth(relativePath, handlers...)
	c.remotePaths[relativePath] = true
	c.paths[relativePath] = true
}
func (c *Context) GetRemoteCheckAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.GetCheckAuth(relativePath, handlers...)
	c.remotePaths[relativePath] = true
	c.paths[relativePath] = true
}
func (c *Context) PostRemote(relativePath string, handlers ...web.HandlerFunc) {
	c.Post(relativePath, handlers...)
	c.remotePaths[relativePath] = true
	c.paths[relativePath] = true
}

func (c *Context) PostRemoteAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.PostAuth(relativePath, handlers...)
	c.remotePaths[relativePath] = true
	c.paths[relativePath] = true
}

func (c *Context) HasPaths(queryPath string) bool {
	_, ok := c.paths[queryPath]
	if ok {
		return ok
	}
	for k, _ := range c.paths {
		h := util.IsMatchPath(queryPath, k)
		if h {
			return h
		}
	}
	return ok
}
func (c *Context) IsRemotePaths(queryPath string) bool {
	_, ok := c.remotePaths[queryPath]
	if ok {
		return ok
	}
	for k, _ := range c.remotePaths {
		h := util.IsMatchPath(queryPath, k)
		if h {
			return h
		}
	}
	return ok
}

// StaticHandle 设置静态文件路由
func (c *Context) StaticHandle(relativePath string, filepath string) {
	c.engine.Use(func(context *gin.Context) {
		path_ := context.Request.URL.Path
		c.log.Debug("StaticHandle", zap.String("Method", context.Request.Method), zap.String("path", path_), zap.Bool("HasPaths", c.HasPaths(path_)))
		if c.HasPaths(path_) || context.Request.Method != "GET" {
			context.Next()
		} else {
			if strings.Contains(path_, "/manifest.json") {
				filePath := path.Join(filepath, "/manifest.json")
				context.File(filePath)
				context.Abort()
			} else {
				relativeFilePath := ""
				if path_ == relativePath {
					relativeFilePath = relativePath + "index.html"
				} else {
					relativeFilePath = path_
				}
				filePath := path.Join(filepath, relativeFilePath)
				context.File(filePath)
				context.Abort()
			}
		}
	})
}

func (c *Context) Go(handle func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				s := string(debug.Stack())
				log.Println(err)
				log.Println(s)
				c.GetLog().Error("Go", zap.Any("err", err), zap.String("info", s))
			}
		}()
		handle()
	}()
}

func (c *Context) IsRemote(context *gin.Context) bool {
	path_ := context.Request.URL.Path
	c.log.Debug("isRemote", zap.Bool("IsRemotePaths", c.IsRemotePaths(path_)), zap.Bool("IsClient", c.serverConfig.IsClient()), zap.Int("ProtoMajor", context.Request.ProtoMajor))
	if c.IsRemotePaths(path_) && c.serverConfig.IsClient() && context.Request.ProtoMajor != 3 {
		return true
	}
	return false
}

func (c *Context) Secret(user, realm string) string {
	c.log.Debug("Secret", zap.String("username", user), zap.String("realm", realm))
	un, code := web.GetUsernameAndCode(user)
	if !c.serverConfig.IsClient() {
		code = ""
	}
	oneUser, err := c.db.GetUserModel().QueryOneUser(un, code)
	if err != nil {
		return ""
	}
	if oneUser != nil {
		v := util.MD5([]byte(fmt.Sprintf("%s:%s:%s", user, realm, oneUser.Password)))
		return v
	}
	return ""
}

func (c *Context) GetReverseProxy(remoteAddress *net.UDPAddr, cert *cert.Certificate) (*khttp.ReverseProxy, error) {
	if cert == nil {
		return c.server.GetReverseProxy(remoteAddress)
	}
	proxy, err := c.server.GetTlsReverseProxy(remoteAddress, cert)
	return proxy, err
}

func (c *Context) ReverseProxy(username, code string, writer http.ResponseWriter, context *gin.Context) {
	certificate, has := c.clientCert.GetCertByCode(username, code)
	if has {
		ds, fa := c.GetDiscoverServer()
		if fa {
			status, err := ds.FindStatusWait(certificate.ServerName, FindWaitTime)
			if err != nil {
				context.AbortWithStatusJSON(200, web.ResponseError(status.GetError().Error()))
			} else {
				if status.IsComplete() {
					reverseProxy, err := c.GetReverseProxy(status.GetAddress(), certificate)
					if err != nil {
						context.AbortWithStatusJSON(200, web.ResponseError(err.Error()))
					} else {
						context.Request.Header.Del("Referer")
						context.Request.Header.Del("Origin")
						reverseProxy.ServeHTTP(writer, context.Request)
					}
				} else {
					context.AbortWithStatusJSON(200, web.ResponseMsg(status.GetCode(), status.GetMsg()))
				}
			}
		}
	} else {
		context.AbortWithStatusJSON(200, web.ResponseError("Incorrect username or password, or certificate not uploaded."))
	}
	context.Abort()
}
func (c *Context) RemoteHandle() {
	c.engine.Use(func(context *gin.Context) {
		c.log.Debug("RemoteHandle", zap.String("RequestURI", context.Request.RequestURI), zap.Bool("isRemote", c.IsRemote(context)))
		if c.IsRemote(context) {
			username := c.digestAuth.ReadAuth(context.Request)
			c.log.Debug("RemoteHandle", zap.String("username", username))
			if username != "" {
				un, code := web.GetUsernameAndCode(username)
				c.log.Debug("RemoteHandle", zap.String("username", un), zap.String("code", code))
				if code == "" {
					context.Next()
				} else {
					c.ReverseProxy(un, code, context.Writer, context)
				}
			}
		}
	})
}
