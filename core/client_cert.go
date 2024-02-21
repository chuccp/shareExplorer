package core

import (
	"errors"
	"github.com/chuccp/kuic/cert"
	"github.com/chuccp/shareExplorer/db"
	"log"
)

type ClientCert struct {
	codeClientCerts []*codeClientCert
	userModel       *db.UserModel
}
type codeClientCert struct {
	clientCertificate *cert.Certificate
	code              string
}

func NewClientCert(userModel *db.UserModel) *ClientCert {
	return &ClientCert{codeClientCerts: make([]*codeClientCert, 0), userModel: userModel}
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
		err := c.LoadCert(user.Username, user.Code, user.CertPath)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *ClientCert) LoadUser(username string, code string) error {
	user, err := c.userModel.QueryOneUser(username, code)
	if err != nil {
		return err
	}
	path := user.CertPath
	return c.LoadCert(username, code, path)
}
func (c *ClientCert) getCert(username string, code string) (*cert.Certificate, bool) {
	for _, certificate := range c.codeClientCerts {
		if certificate.clientCertificate.UserName == username && certificate.code == code {
			return certificate.clientCertificate, true
		}
	}
	return nil, false
}

func (c *ClientCert) getCertByCode(username string, code string) (*cert.Certificate, bool) {
	return c.getCert(username, code)
}

func (c *ClientCert) getCertServername(username string, code string) (string, bool) {
	certs, fa := c.getCert(username, code)
	if fa {
		return certs.ServerName, true
	}
	return "", false
}
func (c *ClientCert) LoadCert(username string, code string, path string) error {
	_, cc, err := cert.ParseClientKuicCertFile(path)
	if err != nil {
		return err
	}
	if username == cc.UserName {
		log.Println("LoadCert username:", username, "ServerName:", cc.ServerName)
		c.codeClientCerts = append(c.codeClientCerts, &codeClientCert{clientCertificate: cc, code: code})
		return nil
	}
	return errors.New("用户名错误")
}
