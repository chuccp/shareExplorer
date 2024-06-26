package core

import (
	"github.com/chuccp/shareExplorer/db"
	"strings"
)

type Config struct {
	IsServer    string
	IsClient    string
	IsNatServer string
}
type ServerConfig struct {
	configModel *db.ConfigModel
	config      *Config
}

func NewServerConfig(configModel *db.ConfigModel) *ServerConfig {
	return &ServerConfig{configModel: configModel, config: &Config{}}
}
func (sc *ServerConfig) Init() ([]*db.Config, error) {
	configs, err := sc.configModel.GetValues("isServer", "isClient", "isNatServer")
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		sc.config.IsServer = ""
		sc.config.IsClient = ""
		sc.config.IsNatServer = ""
	} else {
		for _, config := range configs {
			if config.Key == "isServer" {
				sc.config.IsServer = config.Value
			}
			if config.Key == "isClient" {
				sc.config.IsClient = config.Value
			}
			if config.Key == "isNatServer" {
				sc.config.IsNatServer = config.Value
			}
		}
	}

	return configs, nil
}
func (sc *ServerConfig) GetConfig() *Config {
	return sc.config
}
func (sc *ServerConfig) HasInit() bool {
	return len(sc.config.IsServer) > 0 || len(sc.config.IsClient) > 0 || len(sc.config.IsNatServer) > 0
}
func (sc *ServerConfig) IsServer() bool {
	return strings.Contains(sc.config.IsServer, "true")
}
func (sc *ServerConfig) IsNatServer() bool {
	return strings.Contains(sc.config.IsNatServer, "true")
}
func (sc *ServerConfig) IsClient() bool {
	return strings.Contains(sc.config.IsClient, "true")
}
