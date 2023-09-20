package db

import "database/sql"
import _ "github.com/glebarez/go-sqlite"

type DB struct {
	dbName string
	db     *sql.DB
}

type table struct {
	tableName string
}

func (d *DB) GetConfig() *Config {
	return &Config{table: d.getTable("t_config")}
}

func (d *DB) getTable(tableName string) *table {
	return &table{tableName: tableName}
}
func CreateDb(dbName string) (*DB, error) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return nil, err
	}
	return &DB{db: db, dbName: dbName}, nil
}
