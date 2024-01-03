package core

import (
	db2 "github.com/chuccp/shareExplorer/db"
	"testing"
)

func TestRead(t *testing.T) {
	db, _ := db2.CreateDb("C:\\Users\\cooge\\Documents\\GitHub\\shareExplorer\\share_explorer.db")
	serverConfig := NewServerConfig(db.GetConfigModel())
	serverConfig.Init()

}
