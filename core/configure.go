package core

import "github.com/chuccp/shareExplorer/db"

type Config struct {
	IsServer    string
	IsNatClient string
	IsNatServer string
	HasInit     bool
}
type ServerConfig struct {
	configModel *db.ConfigModel
	config      *Config
}

func NewServerConfig(configModel *db.ConfigModel) *ServerConfig {
	return &ServerConfig{configModel: configModel}
}
func (sc *ServerConfig) Init() error {

	return nil
}
