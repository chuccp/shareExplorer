package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Engine struct {
	*gin.Engine
}

func NewEngine() *Engine {
	return &Engine{Engine: gin.New()}
}
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.Engine.ServeHTTP(w, req)
}
