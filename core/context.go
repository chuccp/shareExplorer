package core

import (
	"github.com/chuccp/kuic/cert"
	khttp "github.com/chuccp/kuic/http"
	"github.com/chuccp/shareExplorer/db"
	"github.com/chuccp/shareExplorer/util"
	"github.com/chuccp/shareExplorer/web"
	"github.com/gin-gonic/gin"
	"log"
	"path"
	"strings"
)

type HandlerFunc func(req *web.Request) (any, error)

type Context struct {
	engine         *gin.Engine
	register       IRegister
	server         *khttp.Server
	discoverServer DiscoverServer
	db             *db.DB
	jwt            *util.Jwt
	paths          map[string]any
	remotePaths    map[string]any
	certManager    *cert.Manager
	serverConfig   *ServerConfig
}

type HandlersChain []HandlerFunc

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
func (c *Context) GetServerConfig() *ServerConfig {
	return c.serverConfig
}
func (c *Context) GetCertManager() *cert.Manager {
	return c.certManager
}
func (c *Context) GetJwt() *util.Jwt {
	return c.jwt
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
func (c *Context) GetHttpClient(address string) (*khttp.Client, error) {
	return c.server.GetHttpClient(address)
}

func (c *Context) Post(relativePath string, handlers ...HandlerFunc) {
	_, ok := c.paths[relativePath]
	if ok {
		return
	}
	c.paths[relativePath] = true
	c.engine.POST(relativePath, c.toGinHandlerFunc(handlers)...)
}
func (c *Context) Get(relativePath string, handlers ...HandlerFunc) {
	_, ok := c.paths[relativePath]
	if ok {
		return
	}
	c.paths[relativePath] = true
	c.engine.GET(relativePath, c.toGinHandlerFunc(handlers)...)
}

func (c *Context) GetRemote(relativePath string, handlers ...HandlerFunc) {
	_, ok := c.paths[relativePath]
	if ok {
		return
	}
	c.Get(relativePath, handlers...)
	c.remotePaths[relativePath] = true
	c.paths[relativePath] = true
}

func (c *Context) PostRemote(relativePath string, handlers ...HandlerFunc) {
	_, ok := c.paths[relativePath]
	if ok {
		return
	}
	c.Post(relativePath, handlers...)
	c.remotePaths[relativePath] = true
	c.paths[relativePath] = true
}

func (c *Context) HasPaths(queryPath string) bool {
	_, ok := c.paths[queryPath]
	return ok
}
func (c *Context) IsRemotePaths(queryPath string) bool {
	_, ok := c.remotePaths[queryPath]
	return ok
}

// StaticHandle 设置静态文件路由
func (c *Context) StaticHandle(relativePath string, filepath string) {
	c.engine.Use(func(context *gin.Context) {
		path_ := context.Request.URL.Path
		if c.HasPaths(path_) {
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
				log.Println(filePath)
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
func (c *Context) GetReverseProxy(remoteAddress string, cert *cert.Certificate) (*khttp.ReverseProxy, error) {
	if cert == nil {
		return c.server.GetReverseProxy(remoteAddress)
	}
	proxy, err := c.server.GetTlsReverseProxy(remoteAddress, cert)
	return proxy, err
}

func (c *Context) RemoteHandle() {
	c.engine.Use(func(context *gin.Context) {
		if c.isRemote(context) {
			ds, fa := c.GetDiscoverServer()
			if fa {
				address, err := ds.FindAddress()
				if err != nil {
					context.AbortWithStatusJSON(200, web.ResponseError(err.Error()))
				} else {
					reverseProxy, err := c.GetReverseProxy(address, nil)
					if err != nil {
						context.AbortWithStatusJSON(200, web.ResponseError(err.Error()))
					} else {

						log.Println("remote", address, context.Request.URL)

						context.Request.Header.Del("Referer")
						context.Request.Header.Del("Origin")
						reverseProxy.ServeHTTP(context.Writer, context.Request)
					}
				}
			}
			context.Abort()
		}
	})
}

func (c *Context) Request(path string, handelFunc func(response *web.ReverseResponse)) error {
	client, err := web.CreateReverseClient(path)
	if err != nil {
		return err
	}
	reverseResponse := client.GetReverseResponse()
	handelFunc(reverseResponse)
	proxy, err := c.server.GetReverseProxy("127.0.0.1:2156")
	if err == nil {
		proxy.ServeHTTP(client.Response, client.Request)
		return nil
	} else {
		return err
	}
}

func (c *Context) toGinHandlerFunc(handlers []HandlerFunc) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = func(context *gin.Context) {
			value, err := handler(web.NewRequest(context, c.jwt))
			if err != nil {
				context.AbortWithStatusJSON(200, web.ResponseError(err.Error()))
			} else {
				switch t := value.(type) {
				case string:
					context.Writer.Write([]byte(t))
				case *web.File:
					context.FileAttachment(t.GetPath(), t.GetFilename())
				default:
					context.AbortWithStatusJSON(200, t)
				}
			}

		}
	}
	return handlerFunc
}
