package db

import (
	"errors"
	"github.com/chuccp/shareExplorer/web"
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username   string    `gorm:"column:username" json:"username"`
	Password   string    `gorm:"column:password" json:"password"`
	Role       string    `gorm:"column:role" json:"role"`
	PathIds    string    `gorm:"column:path_ids" json:"pathIds"`
	CertPath   string    `gorm:"unique;column:cert_path" json:"certPath"`
	Code       string    `gorm:"unique;column:code" json:"code"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime"`
}

type UserModel struct {
	db        *gorm.DB
	tableName string
}

func (u *UserModel) DeleteTable() error {
	if !u.IsExist() {
		return nil
	}
	tx := u.db.Table(u.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{})
	return tx.Error
}
func (u *UserModel) deleteTable() error {
	tx := u.db.Table(u.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{})
	return tx.Error
}
func (u *UserModel) IsExist() bool {
	return u.db.Migrator().HasTable(u.tableName)
}

func (u *UserModel) HasData() bool {
	var num int64
	u.db.Table(u.tableName).Count(&num)
	return num > 0
}

func (u *UserModel) createTable() error {
	err := u.db.Table(u.tableName).AutoMigrate(&User{})
	return err
}
func (u *UserModel) NewModel(db *gorm.DB) *UserModel {
	return &UserModel{db: db, tableName: u.tableName}
}
func (u *UserModel) AddUser(username string, password string, role string, path string) error {
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
		Code:       username,
		CertPath:   path,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	})
	return tx.Error
}

func (u *UserModel) DeleteUser(username string) error {
	var v User
	tx := u.db.Table(u.tableName).Where("`username` = ?", username).First(&v)
	if tx.Error != nil {
		return tx.Error
	}
	if v.Role == "admin" {
		return errors.New("管理员账号不能删除")
	}
	tx = u.db.Table(u.tableName).Where("`username` = ? and role!=?", username, "admin").Delete(&User{})
	if tx.Error != nil {
		return tx.Error
	}
	return tx.Error
}
func (u *UserModel) EditUser(id uint, username string, password string, pathIds string) error {
	tx := u.db.Table(u.tableName).Where(&User{
		Id: id,
	}).Updates(&User{
		Username:   username,
		Password:   password,
		PathIds:    pathIds,
		UpdateTime: time.Now(),
	})
	return tx.Error
}

func (u *UserModel) AddGuestUser(username string, password string, pathIds string, path string) error {
	if !u.IsExist() {
		err := u.createTable()
		if err != nil {
			return err
		}
	}
	tx := u.db.Table(u.tableName).Create(&User{
		Username:   username,
		Password:   password,
		Role:       "guest",
		Code:       username,
		PathIds:    pathIds,
		CertPath:   path,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	})
	return tx.Error
}
func (u *UserModel) AddClientUser(username string, code string, certPath string) error {
	if !u.IsExist() {
		err := u.createTable()
		if err != nil {
			return err
		}
	}

	var count2 int64
	tx := u.db.Table(u.tableName).Where("code=?  limit 1", code).Count(&count2)
	if tx.Error != nil {
		return tx.Error
	}
	if count2 > 0 {
		return errors.New("code已经存在")
	}
	var count int64
	tx = u.db.Table(u.tableName).Where("cert_path=?  limit 1", certPath).Count(&count)
	if tx.Error != nil {
		return tx.Error
	}
	if count > 0 {
		return errors.New("证书已经存在")
	}
	tx = u.db.Table(u.tableName).Create(&User{
		Username:   username,
		Password:   "",
		Role:       "client",
		CertPath:   certPath,
		PathIds:    "",
		Code:       code,
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
func (u *UserModel) QueryOneUser(username string, code string) (*User, error) {
	if !u.IsExist() {
		err := u.createTable()
		if err != nil {
			return nil, err
		}
	}
	var user User
	tx := u.db.Table(u.tableName).Find(&user, "username=? and code=? limit 1", username, code)
	if tx.Error == nil {
		return &user, nil
	}
	return nil, tx.Error
}
func (u *UserModel) QueryAllUser() ([]*User, error) {
	if !u.IsExist() {
		err := u.createTable()
		if err != nil {
			return nil, err
		}
	}
	var users01 []*User
	tx := u.db.Table(u.tableName).Find(&users01)
	if tx.Error == nil {
		return users01, nil
	}
	return nil, tx.Error
}
func (u *UserModel) QueryPage(pageNo int, pageSize int) ([]*User, int64, error) {

	users := make([]*User, 0)
	if !u.IsExist() {
		return users, 0, nil
	}
	var users01 []*User
	tx := u.db.Table(u.tableName).Order("`id` desc").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&users01)
	if tx.Error == nil {
		var num int64
		tx = u.db.Table(u.tableName).Count(&num)
		if tx.Error == nil {
			return users01, num, nil
		}
	}
	return nil, 0, tx.Error
}
func (u *UserModel) QueryById(id uint) (*User, error) {
	if !u.IsExist() {
		return nil, web.NotFound
	}
	var users01 User
	tx := u.db.Table(u.tableName).Where(&User{Id: id}).First(&users01)
	if tx.Error == nil {
		return &users01, nil
	}
	return nil, tx.Error
}
