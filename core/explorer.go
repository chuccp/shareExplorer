package core

import (
	"container/list"
	"github.com/chuccp/utils/config"
)

type ShareExplorer struct {
	ctx *Context
	list  *list.List
}
func (se *ShareExplorer)AddServer(server Server){
	se.list.PushBack(server)
}
func (se *ShareExplorer) Start() {
	for ele := se.list.Front();ele!=nil ; ele = ele.Next() {
		server:=(ele.Value).(Server)
		server.Init(se.ctx)
		go  server.Start()
	}
}
func NewShareExplorer(cfg *config.Config)*ShareExplorer  {
	return &ShareExplorer{ctx:NewContext(cfg),list:list.New()}
}