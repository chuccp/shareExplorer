package core

import "github.com/chuccp/cokePush/config"

type ShareExplorer struct {
	config *config.Config
}

func (ShareExplorer *ShareExplorer) Start()error {

	return nil
}
func NewShareExplorer(config *config.Config)*ShareExplorer  {
	return &ShareExplorer{config: config}
}