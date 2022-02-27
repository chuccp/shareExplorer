package core

import "github.com/chuccp/cokePush/config"

type Server interface {
	Start() error
	Init(*config.Config)
}
