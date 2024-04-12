package db

import (
	"database/sql"
	"github.com/chuccp/shareExplorer/util"
	"log"
	"testing"
)
import _ "github.com/glebarez/go-sqlite"

func TestDB(t *testing.T) {
	db, err := sql.Open("sqlite", "share_explorer.db")
	if err != nil {
		log.Println(err)
		return
	}
	query, err := db.Query("SELECT COUNT(*) as num FROM sqlite_master WHERE type = 'table' AND name ='t_config'")
	if err != nil {
		log.Println(err)
		return
	}
	var num int64
	query.Scan(&num)

	println(num)

}
func TestCreate(t *testing.T) {
	//db, err := sql.Open("sqlite", "share_explorer.db")
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//query, err := db.Exec("")
	//println(query, err)

	t.Log(util.SplitPath("/aaa/"))

}
