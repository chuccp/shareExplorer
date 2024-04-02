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
	"net"
	"path"
	"strings"
)

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

func (c *Context) isRemote(context *gin.Context) bool {
	path_ := context.Request.URL.Path
	if c.IsRemotePaths(path_) && !c.serverConfig.IsServer() && context.Request.ProtoMajor != 3 {
		return true
	}
	return false
}

func (c *Context) Secret(user, realm string) string {
	if user == "111111" {
		v := util.MD5([]byte(fmt.Sprintf("%s:%s:%s", user, realm, "111111")))
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

func (c *Context) RemoteHandle() {
	c.engine.Use(func(context *gin.Context) {
		if c.isRemote(context) {
			username := context.Request.FormValue("username")
			if username == "" {
				username = context.Request.Header.Get("Username")
			}
			code := context.Request.FormValue("code")
			if code == "" {
				code = context.Request.Header.Get("Code")
			}
			isStart := context.Request.FormValue("start")
			certificate, has := c.clientCert.getCertByCode(username, code)
			if has {
				ds, fa := c.GetDiscoverServer()
				if fa {
					status := ds.FindStatus(certificate.ServerName, strings.Contains(isStart, "true"))
					if status.GetError() != nil {
						context.AbortWithStatusJSON(200, web.ResponseError(status.GetError().Error()))
					} else {
						if status.IsComplete() {
							reverseProxy, err := c.GetReverseProxy(status.GetAddress(), certificate)
							if err != nil {
								context.AbortWithStatusJSON(200, web.ResponseError(err.Error()))
							} else {
								context.Request.Header.Del("Referer")
								context.Request.Header.Del("Origin")
								reverseProxy.ServeHTTP(context.Writer, context.Request)
							}
						} else {
							context.AbortWithStatusJSON(200, web.ResponseMsg(status.GetCode(), status.GetMsg()))
						}
					}
				}
			} else {
				context.AbortWithStatusJSON(200, web.ResponseError("用户名有误或未上传证书"))
			}
			context.Abort()
		}
	})
}
