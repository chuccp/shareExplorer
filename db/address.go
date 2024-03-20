package db

import (
	"github.com/chuccp/shareExplorer/util"
	"gorm.io/gorm"
	"time"
)

type Address struct {
	Id           uint      `gorm:"primaryKey;autoIncrement;column:id"`
	Address      string    `gorm:"unique;column:address"`
	FailNum      string    `gorm:"column:fail_num"`
	ServerName   string    `gorm:"unique;column:serverName"`
	LastFailTime time.Time `gorm:"column:last_fail_time"`
	CreateTime   time.Time `gorm:"column:create_time"`
	UpdateTime   time.Time `gorm:"column:update_time"`
}
type AddressModel struct {
	db        *gorm.DB
	tableName string
}

func (a *AddressModel) NewModel(db *gorm.DB) *AddressModel {
	return &AddressModel{db: db, tableName: a.tableName}
}

func (a *AddressModel) IsExist() bool {
	return a.db.Migrator().HasTable(a.tableName)
}
func (a *AddressModel) createTable() error {
	err := a.db.Table(a.tableName).AutoMigrate(&Address{})
	return err
}

func (a *AddressModel) DeleteTable() error {
	if !a.IsExist() {
		return nil
	}
	tx := a.db.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Address{})
	return tx.Error
}

func (a *AddressModel) QueryAddresses() ([]*Address, error) {
	var addr []*Address
	tx := a.db.Table(a.tableName).Find(&addr)
	return addr, tx.Error
}

func (a *AddressModel) UpdateServerNameByAddress(address, serverName string) error {
	tx := a.db.Table(a.tableName).Where(&Address{
		Address: address,
	}).Updates(&Address{
		ServerName: serverName,
		UpdateTime: time.Now(),
	})
	return tx.Error
}

func (a *AddressModel) AddAddress(addresses []string) error {
	if !a.IsExist() {
		err := a.createTable()
		if err != nil {
			return err
		}
	}
	var addr []*Address
	tx := a.db.Table(a.tableName).Where(" address IN ?", addresses).Find(&addr)
	if tx.Error != nil {
		return tx.Error
	}
	for _, add := range addr {
		addresses = util.DeleteElement(addresses, add.Address)
	}
	var adds = make([]*Address, len(addresses))
	for index, address := range addresses {
		adds[index] = &Address{
			Address:    address,
			CreateTime: time.Now(),
			UpdateTime: time.Now()}
	}
	tx = a.db.Table(a.tableName).CreateInBatches(adds, len(adds))
	return tx.Error
}
