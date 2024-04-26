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
	Seed         int       `gorm:"column:seed"`
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

func (a *AddressModel) isExist() bool {
	return a.db.Migrator().HasTable(a.tableName)
}
func (a *AddressModel) CreateTable() error {
	if a.isExist() {
		return nil
	}
	err := a.db.Table(a.tableName).AutoMigrate(&Address{})
	return err
}

func (a *AddressModel) DeleteTable() error {
	tx := a.db.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Address{})
	return tx.Error
}

func (a *AddressModel) QueryAddresses() ([]*Address, error) {
	var addr []*Address
	tx := a.db.Table(a.tableName).Find(&addr).Limit(272)
	return addr, tx.Error
}

func (a *AddressModel) QuerySeedAddresses() ([]*Address, error) {
	var addr []*Address
	tx := a.db.Table(a.tableName).Where(&Address{
		Seed: 1,
	}).Find(&addr)
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
func (a *AddressModel) UpdateAddressByServerName(serverName, address string) error {
	tx := a.db.Table(a.tableName).Where(&Address{
		ServerName: serverName,
	}).Updates(&Address{
		Address:    address,
		UpdateTime: time.Now(),
	})
	return tx.Error
}

func (a *AddressModel) AddAddresses(addresses []string, isSeed bool) error {
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
		addr := &Address{
			Address:    address,
			CreateTime: time.Now(),
			UpdateTime: time.Now(), Seed: 0}
		if isSeed {
			addr.Seed = 1
		}
		adds[index] = addr
	}
	tx = a.db.Table(a.tableName).CreateInBatches(adds, len(adds))
	return tx.Error
}
func (a *AddressModel) AddAddress(address string, serverName string, isSeed bool) error {
	var adds []*Address
	tx := a.db.Table(a.tableName).Where(&Address{
		Address: address,
	}).Find(&adds)
	if tx.Error != nil {
		return tx.Error
	}
	if len(adds) > 0 {
		return nil
	}
	addr := &Address{
		Address:    address,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		ServerName: serverName,
		Seed:       0}

	if isSeed {
		addr.Seed = 1
	}
	tx = a.db.Table(a.tableName).Create(addr)
	return tx.Error
}
