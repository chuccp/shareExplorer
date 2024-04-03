package db

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Path struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name       string    `gorm:"unique;column:name" json:"name"`
	Path       string    `gorm:"unique;column:path" json:"path"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime"`
}

var pathMap = NewMap[*Path]()

type PathModel struct {
	db        *gorm.DB
	tableName string
}

func (a *PathModel) isExist() bool {
	return a.db.Migrator().HasTable(a.tableName)
}

func (a *PathModel) DeleteTable() error {
	pathMap.Clean()
	tx := a.db.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Path{})
	return tx.Error
}

func (a *PathModel) NewModel(db *gorm.DB) *PathModel {
	return &PathModel{db: db, tableName: a.tableName}
}

func (a *PathModel) CreateTable() error {
	if a.isExist() {
		return nil
	}
	err := a.db.Table(a.tableName).AutoMigrate(&Path{})
	return err
}

func (a *PathModel) Create(name string, path string) error {

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

func (a *PathModel) Query(name string) (*Path, error) {
	var path Path
	p, ok := pathMap.Get(name)
	if ok {
		return p, nil
	}
	tx := a.db.Table(a.tableName).Where(" `name` = ? ", name).First(&path)
	if tx.Error != nil {
		return nil, tx.Error
	}
	pathMap.Save(name, &path)
	return &path, nil
}

func (a *PathModel) Update(id int, name string, path string) error {
	pathMap.Clean()
	tx := a.db.Table(a.tableName).Where(&Path{Id: uint(id)}).Updates(&Path{Name: name, Path: path, UpdateTime: time.Now()})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
func (a *PathModel) Delete(id uint) error {
	pathMap.Clean()
	tx := a.db.Table(a.tableName).Where("`id` = ?", id).Delete(&Path{})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (a *PathModel) QueryPage(pageNo int, pageSize int) ([]*Path, int64, error) {
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
func (a *PathModel) QueryAll() ([]*Path, error) {
	var paths01 []*Path
	tx := a.db.Table(a.tableName).Find(&paths01)
	if tx.Error == nil {
		return paths01, nil
	}
	return nil, tx.Error
}
func (a *PathModel) QueryById(id uint) (*Path, error) {
	var users01 Path
	tx := a.db.Table(a.tableName).Where(&Path{Id: id}).First(&users01)
	if tx.Error == nil {
		return &users01, nil
	}
	return nil, tx.Error
}
