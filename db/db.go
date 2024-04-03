package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	dbName string
	db     *gorm.DB
}

func (d *DB) GetConfigModel() *ConfigModel {
	return &ConfigModel{db: d.db, tableName: "t_config"}
}

func (d *DB) GetUserModel() *UserModel {
	return &UserModel{db: d.db, tableName: "t_user"}
}
func (d *DB) GetAddressModel() *AddressModel {
	return &AddressModel{db: d.db, tableName: "t_address"}
}

func (d *DB) GetPathModel() *PathModel {
	return &PathModel{db: d.db, tableName: "t_path"}
}

func (d *DB) GetRawDB() *gorm.DB {
	return d.db
}
func (d *DB) Reset() error {
	err := d.db.Transaction(func(tx *gorm.DB) error {
		err := d.GetConfigModel().NewModel(tx).DeleteTable()
		if err != nil {
			return err
		}
		err = d.GetPathModel().NewModel(tx).DeleteTable()
		if err != nil {
			return err
		}
		err = d.GetAddressModel().NewModel(tx).DeleteTable()
		if err != nil {
			return err
		}
		err = d.GetUserModel().NewModel(tx).DeleteTable()
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func CreateDb(dbName string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}
	dbTable := &DB{db: db, dbName: dbName}
	err = dbTable.GetPathModel().CreateTable()
	if err != nil {
		return nil, err
	}
	err = dbTable.GetUserModel().CreateTable()
	if err != nil {
		return nil, err
	}
	err = dbTable.GetAddressModel().CreateTable()
	if err != nil {
		return nil, err
	}
	err = dbTable.GetConfigModel().CreateTable()
	return dbTable, err
}
