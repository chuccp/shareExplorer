package core

import (
	"github.com/chuccp/shareExplorer/db"
	"strings"
)

type Config struct {
	IsServer    string
	IsNatClient string
	IsNatServer string
}
type ServerConfig struct {
	configModel *db.ConfigModel
	config      *Config
}

func NewServerConfig(configModel *db.ConfigModel) *ServerConfig {
	return &ServerConfig{configModel: configModel, config: &Config{}}
}
func (sc *ServerConfig) Init() error {
	configs, err := sc.configModel.GetValues("isServer", "isNatClient", "isNatServer")
	if err != nil {
		return err
	}
	for _, config := range configs {
		if config.Key == "isServer" {
			sc.config.IsServer = config.Value
		}
		if config.Key == "isNatClient" {
			sc.config.IsNatClient = config.Value
		}
		if config.Key == "isNatServer" {
			sc.config.IsNatServer = config.Value
		}
	}
	return nil
}
func (sc *ServerConfig) HasInit() bool {
	return len(sc.config.IsServer) > 0
}
func (sc *ServerConfig) IsServer() bool {
	return strings.Contains(sc.config.IsServer, "true")
}
func (sc *ServerConfig) IsNatServer() bool {
	return strings.Contains(sc.config.IsNatServer, "true")
}
func (sc *ServerConfig) IsNatClient() bool {
	return strings.Contains(sc.config.IsNatClient, "true")
}
