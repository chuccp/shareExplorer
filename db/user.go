package db

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id"`
	Username   string    `gorm:"unique;column:username"`
	Password   string    `gorm:"column:password"`
	Role       string    `gorm:"column:role"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

type UserModel struct {
	db        *gorm.DB
	tableName string
}

func (u *UserModel) IsExist() bool {
	return u.db.Migrator().HasTable(u.tableName)
}
func (u *UserModel) create() error {
	err := u.db.Table(u.tableName).AutoMigrate(&User{})
	return err
}
func (u *UserModel) NewModel(db *gorm.DB) *UserModel {
	return &UserModel{db: db, tableName: u.tableName}
}
func (u *UserModel) AddUser(username string, password string, role string) error {
	if !u.IsExist() {
		err := u.create()
		if err != nil {
			return err
		}
	}
	tx := u.db.Table(u.tableName).Create(&User{
		Username:   username,
		Password:   password,
		Role:       role,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	})
	return tx.Error
}
