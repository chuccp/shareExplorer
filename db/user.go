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
func (u *UserModel) createTable() error {
	err := u.db.Table(u.tableName).AutoMigrate(&User{})
	return err
}
func (u *UserModel) NewModel(db *gorm.DB) *UserModel {
	return &UserModel{db: db, tableName: u.tableName}
}
func (u *UserModel) AddUser(username string, password string, role string) error {
	if !u.IsExist() {
		err := u.createTable()
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
func (u *UserModel) QueryUser(username string, password string) (*User, error) {
	if !u.IsExist() {
		err := u.createTable()
		if err != nil {
			return nil, err
		}
	}
	var user User
	tx := u.db.Table(u.tableName).Find(&user, "username=? and password=? limit 1", username, password)
	if tx.Error == nil {
		return &user, nil
	}
	return nil, tx.Error
}
