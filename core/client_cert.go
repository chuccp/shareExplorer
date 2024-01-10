package core

import (
	"errors"
	"github.com/chuccp/kuic/cert"
	"github.com/chuccp/shareExplorer/db"
	"log"
)

type ClientCert struct {
	clientCertificates []*cert.Certificate
	userModel          *db.UserModel
}

func NewClientCert(userModel *db.UserModel) *ClientCert {
	return &ClientCert{clientCertificates: make([]*cert.Certificate, 0), userModel: userModel}
}
func (c *ClientCert) init() error {
	err := c.loadAllUser()
	if err != nil {
		return err
	}
	return nil
}
func (c *ClientCert) loadAllUser() error {
	users, err := c.userModel.QueryAllUser()
	if err != nil {
		return err
	}
	for _, user := range users {
		log.Println("loadAllUser username:", user.Username, "Path:", user.CertPath)
		c.LoadCert(user.Username, user.CertPath)
	}
	return nil
}
func (c *ClientCert) LoadUser(username string) error {
	user, err := c.userModel.QueryOneUser(username)
	if err != nil {
		return err
	}
	path := user.CertPath
	return c.LoadCert(username, path)
}
func (c *ClientCert) getCert(username string) (*cert.Certificate, bool) {
	for _, certificate := range c.clientCertificates {
		if certificate.UserName == username {
			return certificate, true
		}
	}
	return nil, false
}
func (c *ClientCert) LoadCert(username string, path string) error {
	_, cc, err := cert.ParseClientKuicCertFile(path)
	if err != nil {
		return err
	}
	if username == cc.UserName {
		c.clientCertificates = append(c.clientCertificates, cc)
		return nil
	}
	return errors.New("用户名错误")
}
