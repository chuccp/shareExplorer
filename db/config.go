package db

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Config struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id"`
	Key        string    `gorm:"unique;column:key"`
	Value      string    `gorm:"column:value"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

type ConfigModel struct {
	db        *gorm.DB
	tableName string
}

func (u *ConfigModel) isExist() bool {
	return u.db.Migrator().HasTable(u.tableName)
}
func (u *ConfigModel) CreateTable() error {

	if u.isExist() {
		return nil
	}

	err := u.db.Table(u.tableName).AutoMigrate(&Config{})
	return err
}

func (u *ConfigModel) DeleteTable() error {
	tx := u.db.Table(u.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Config{})
	return tx.Error
}

func (u *ConfigModel) HasData() bool {
	var num int64
	u.db.Table(u.tableName).Where(" `key` = 'isServer' ").Count(&num)
	return num > 0
}

func (u *ConfigModel) GetValue(key string) (string, bool) {
	var config Config
	tx := u.db.Table(u.tableName).Where(&Config{Key: key}).First(&config)
	return config.Value, !(tx.Error != nil || len(config.Value) == 0)
}

func (u *ConfigModel) GetValues(keys ...string) ([]*Config, error) {

	var configs []*Config
	tx := u.db.Table(u.tableName).Where("`key` in ?", keys).Find(&configs)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return configs, nil
	}
	return configs, tx.Error
}

func (u *ConfigModel) NewModel(db *gorm.DB) *ConfigModel {
	return &ConfigModel{db: db, tableName: u.tableName}
}
func (u *ConfigModel) Create(key string, value string) error {

	rows, err := u.db.Table(u.tableName).Where(" `key` = ?", key).Rows()
	if err != nil {
		return err
	}
	if rows.Next() {
		tx := u.db.Table(u.tableName).Where(" `key` = ?", key).Updates(&Config{
			Key:        key,
			Value:      value,
			UpdateTime: time.Now()})
		return tx.Error
	}
	tx := u.db.Table(u.tableName).Create(&Config{
		Key:        key,
		Value:      value,
		CreateTime: time.Now(),
		UpdateTime: time.Now()})
	return tx.Error
}
