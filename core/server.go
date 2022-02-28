package core

type Server interface {
	Start()
	Init(ctx *Context)
}
