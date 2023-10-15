package db

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Path struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id"`
	Name       string    `gorm:"unique;column:name"`
	Path       string    `gorm:"unique;column:path"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}
type PathModel struct {
	db        *gorm.DB
	tableName string
}

func (a *PathModel) IsExist() bool {
	return a.db.Migrator().HasTable(a.tableName)
}
func (a *PathModel) createTable() error {
	err := a.db.Table(a.tableName).AutoMigrate(&Path{})
	return err
}
func (a *PathModel) Create(name string, path string) error {
	if !a.IsExist() {
		a.createTable()
	}
	rows, err := a.db.Table(a.tableName).Where(" `name` = ? or `path`=?", name, path).Rows()
	if err != nil {
		return err
	}
	if rows.Next() {
		return errors.New("存在重复")
	}
	tx := a.db.Table(a.tableName).Create(&Path{
		Name:       name,
		Path:       path,
		CreateTime: time.Now(),
		UpdateTime: time.Now()})
	return tx.Error
}
