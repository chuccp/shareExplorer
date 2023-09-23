package db

import (
	"gorm.io/gorm"
	"time"
)

type Config struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id"`
	Key        string    `gorm:"unique;column:username"`
	Value      string    `gorm:"column:value"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

type ConfigModel struct {
	db        *gorm.DB
	tableName string
}

func (u *ConfigModel) IsExist() bool {
	return u.db.Migrator().HasTable(u.tableName)
}
func (u *ConfigModel) create() error {
	err := u.db.Table(u.tableName).AutoMigrate(&Config{})
	return err
}
func (u *ConfigModel) NewModel(db *gorm.DB) *ConfigModel {
	return &ConfigModel{db: db, tableName: u.tableName}
}
func (u *ConfigModel) Create(key string, value string) error {

	if !u.IsExist() {
		u.create()
	}
	tx := u.db.Table(u.tableName).Create(&Config{
		Key:        key,
		Value:      value,
		CreateTime: time.Now(),
		UpdateTime: time.Now()})
	return tx.Error
}
