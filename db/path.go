package db

import (
	"errors"
	"github.com/chuccp/shareExplorer/web"
	"gorm.io/gorm"
	"log"
	"time"
)

type Path struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name       string    `gorm:"unique;column:name" json:"name"`
	Path       string    `gorm:"unique;column:path" json:"path"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime"`
}
type PathModel struct {
	db        *gorm.DB
	tableName string
}

func (a *PathModel) IsExist() bool {

	return a.db.Migrator().HasTable(a.tableName)

}

func (a *PathModel) DeleteTable() error {

	if !a.IsExist() {
		return nil
	}
	log.Println("PathModel")
	tx := a.db.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Path{})
	return tx.Error
}

func (a *PathModel) NewModel(db *gorm.DB) *PathModel {
	return &PathModel{db: db, tableName: a.tableName}
}

func (a *PathModel) createTable() error {
	err := a.db.Table(a.tableName).AutoMigrate(&Path{})
	return err
}
func (a *PathModel) Create(name string, path string) error {

	if !a.IsExist() {
		a.createTable()
	}
	var num int64
	tx := a.db.Table(a.tableName).Where(" `name` = ? or `path`=?", name, path).Count(&num)
	if tx.Error != nil {
		return tx.Error
	}
	if num > 0 {
		return errors.New("存在重复")
	}
	tx = a.db.Table(a.tableName).Create(&Path{
		Name:       name,
		Path:       path,
		CreateTime: time.Now(),
		UpdateTime: time.Now()})
	return tx.Error
}
func (a *PathModel) Update(id int, name string, path string) error {

	if !a.IsExist() {
		a.createTable()
	}
	tx := a.db.Table(a.tableName).Where(&Path{Id: uint(id)}).Updates(&Path{Name: name, Path: path, UpdateTime: time.Now()})
	return tx.Error
}
func (a *PathModel) Delete(id uint) error {
	if !a.IsExist() {
		return nil
	}
	tx := a.db.Table(a.tableName).Where("`id` = ?", id).Delete(&Path{})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (a *PathModel) QueryPage(pageNo int, pageSize int) ([]*Path, int64, error) {

	paths := make([]*Path, 0)
	if !a.IsExist() {
		return paths, 0, nil
	}
	var paths01 []*Path
	tx := a.db.Table(a.tableName).Order("`id` desc").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&paths01)
	if tx.Error == nil {
		var num int64
		tx = a.db.Table(a.tableName).Count(&num)
		if tx.Error == nil {
			return paths01, num, nil
		}
	}
	return nil, 0, tx.Error
}
func (a *PathModel) QueryById(id uint) (*Path, error) {
	if !a.IsExist() {
		return nil, web.NotFound
	}
	var users01 Path
	tx := a.db.Table(a.tableName).Where(&Path{Id: id}).First(&users01)
	if tx.Error == nil {
		return &users01, nil
	}
	return nil, tx.Error
}
