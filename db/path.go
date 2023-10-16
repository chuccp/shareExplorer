package db

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Path struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id"`
	Name       string    `gorm:"unique;column:name" json:"name"`
	Path       string    `gorm:"unique;column:path" json:"path"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime"`
}
type PathModel struct {
	db        *gorm.DB
	tableName string
}

func (a *PathModel) IsExist() (bool, error) {
	var num int64
	tx := a.db.Table("sqlite_master").Where("type='table' AND name=?", a.tableName).Count(&num)
	return num > 0, tx.Error

}
func (a *PathModel) createTable() error {
	err := a.db.Table(a.tableName).AutoMigrate(&Path{})
	return err
}
func (a *PathModel) Create(name string, path string) error {
	exist, err := a.IsExist()
	if err != nil {
		return err
	}
	if !exist {
		a.createTable()
	}
	rows, err := a.db.Table(a.tableName).Where(" `name` = ? or `path`=?", name, path).Rows()
	if err != nil {
		return err
	}
	if rows.Next() {
		rows.Close()
		return errors.New("存在重复")
	}
	tx := a.db.Table(a.tableName).Create(&Path{
		Name:       name,
		Path:       path,
		CreateTime: time.Now(),
		UpdateTime: time.Now()})
	return tx.Error
}

func (a *PathModel) QueryPage(pageNo int, pageSize int) ([]*Path, int64, error) {

	exist, err := a.IsExist()
	if err != nil {
		return nil, 0, err
	}

	paths := make([]*Path, 0)
	if !exist {
		return paths, 0, nil
	}
	var paths01 []*Path
	tx := a.db.Table(a.tableName).Offset(pageNo * pageSize).Limit(pageSize).Find(&paths01)
	if tx.Error == nil {
		var num int64
		tx = a.db.Table(a.tableName).Count(&num)
		if tx.Error == nil {
			return paths01, num, nil
		}
	}
	return nil, 0, tx.Error
}
