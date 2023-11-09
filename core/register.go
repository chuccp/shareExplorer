package core

import (
	"github.com/chuccp/shareExplorer/util"
	"sync"
)

type Register struct {
	servers *sync.Map
	cfg     *util.Config
}

func NewRegister(cfg *util.Config) *Register {
	return &Register{cfg: cfg, servers: new(sync.Map)}
}
func (register *Register) AddServer(server Server) {
	register.servers.Store(server.GetName(), server)
}
func (register *Register) GetConfig() *util.Config {
	return register.cfg
}
func (register *Register) Range(f func(server Server) bool) {
	register.servers.Range(func(_, value any) bool {
		return f(value.(Server))
	})
}

func (register *Register) Create() (*ShareExplorer, error) {
	return CreateShareExplorer(register)
}
