package core

import "github.com/chuccp/utils/config"

type Context struct {
	cfg *config.Config
}

func (ctx *Context) GetConfig()*config.Config  {
	return ctx.cfg
}
func NewContext(cfg *config.Config) *Context {
	return &Context{cfg: cfg}
}