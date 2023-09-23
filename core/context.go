package core

import (
	khttp "github.com/chuccp/kuic/http"
	"github.com/chuccp/shareExplorer/db"
	"github.com/chuccp/shareExplorer/util"
	"github.com/chuccp/shareExplorer/web"
	"github.com/gin-gonic/gin"
	"strings"
)

type HandlerFunc func(req *web.Request) (any, error)

type Context struct {
	engine    *gin.Engine
	register  IRegister
	server    *khttp.Server
	traversal Traversal
	db        *db.DB
}

type HandlersChain []HandlerFunc

func (c *Context) Get(relativePath string, handlers ...HandlerFunc) {
	c.engine.GET(relativePath, c.toGinHandlerFunc(handlers...)...)
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
func (c *Context) SetTraversal(traversal Traversal) {
	c.traversal = traversal
}
func (c *Context) getTraversal() (Traversal, bool) {
	return c.traversal, c.traversal != nil
}
func (c *Context) GetConfigInt(section, name string) (int, error) {
	return c.register.GetConfig().GetInt(section, name)
}
func (c *Context) GetHttpClient(address string) (*khttp.Client, error) {
	return c.server.GetHttpClient(address)
}
func (c *Context) toGinHandlerFunc(handlers ...HandlerFunc) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = func(context *gin.Context) {
			value, err := handler(web.NewRequest(context))
			switch t := value.(type) {
			case string:
				if err != nil {
					context.AbortWithError(500, err)
				} else {
					context.Writer.Write([]byte(t))
				}
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

func (c *Context) Post(relativePath string, handlers ...HandlerFunc) {
	c.engine.POST(relativePath, c.toGinHandlerFunc(handlers...)...)
}
