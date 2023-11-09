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
	engine       *gin.Engine
	register     IRegister
	server       *khttp.Server
	traversal    TraversalServer
	db           *db.DB
	jwt          *util.Jwt
	paths        map[string]any
	remotePaths  map[string]any
	certManager  *cert.Manager
	serverConfig *ServerConfig
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
func (c *Context) GetCertManager() *cert.Manager {
	return c.certManager
}
func (c *Context) GetJwt() *util.Jwt {
	return c.jwt
}
func (c *Context) SetTraversal(traversal TraversalServer) {
	c.traversal = traversal
}
func (c *Context) GetTraversal() (TraversalServer, bool) {
	return c.traversal, c.traversal != nil
}
func (c *Context) GetConfigInt(section, name string) (int, error) {
	return c.register.GetConfig().GetInt(section, name)
}
func (c *Context) GetHttpClient(address string) (*khttp.Client, error) {
	return c.server.GetHttpClient(address)
}

func (c *Context) Post(relativePath string, handlers ...HandlerFunc) {
	c.paths[relativePath] = true
	c.engine.POST(relativePath, c.toGinHandlerFunc(handlers)...)
}
func (c *Context) Get(relativePath string, handlers ...HandlerFunc) {
	c.paths[relativePath] = true
	c.engine.GET(relativePath, c.toGinHandlerFunc(handlers)...)
}

func (c *Context) GetRemote(relativePath string, handlers ...HandlerFunc) {
	c.Get(relativePath, handlers...)
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
		log.Println("==StaticHandle===:", path_, c.HasPaths(path_), c.IsRemotePaths(path_), "----:", c.serverConfig.IsServer())
		if c.HasPaths(path_) {
			context.Next()
		} else {
			if strings.Contains(path_, "/manifest.json") {
				filePath := path.Join(filepath, "/manifest.json")
				context.File(filePath)
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
			}
		}
	})
}
func (c *Context) RemoteHandle() {
	log.Println("()()()()()()()")
	c.engine.Use(func(context *gin.Context) {
		log.Println("===============================")
		path_ := context.Request.URL.Path
		log.Println("=RemoteHandle===:", c.IsRemotePaths(path_), "----:", c.serverConfig.IsServer(), context.Request.ProtoMajor, context.Request.Proto, context.Request.ProtoMinor)
		if c.IsRemotePaths(path_) && c.serverConfig.IsServer() && context.Request.ProtoMajor != 3 {
			proxy, err := c.server.GetReverseProxy("127.0.0.1:2156")
			log.Println(proxy, err)
			if err == nil {
				context.Request.Header.Del("Origin")
				context.Request.Header.Del("Referer")
				proxy.ServeHTTP(context.Writer, context.Request)
			}
		}
	})
}

func (c *Context) toGinHandlerFunc(handlers []HandlerFunc) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = func(context *gin.Context) {
			value, err := handler(web.NewRequest(context, c.jwt))
			switch t := value.(type) {
			case string:
				if err != nil {
					context.AbortWithError(500, err)
				} else {
					context.Writer.Write([]byte(t))
				}
			case *web.File:
				context.FileAttachment(t.GetPath(), t.GetFilename())
			default:
				if err != nil {
					if t != nil {
						context.AbortWithStatusJSON(500, t)
					} else {
						context.AbortWithError(500, err)
					}
				} else {
					context.AbortWithStatusJSON(200, t)
				}
			}
		}
	}
	return handlerFunc
}
