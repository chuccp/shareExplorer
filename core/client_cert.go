package core

import (
	"errors"
	"github.com/chuccp/kuic/cert"
	"github.com/chuccp/shareExplorer/db"
	"go.uber.org/zap"
)

type ClientCert struct {
	codeClientCerts []*CodeClientCert
	userModel       *db.UserModel
	context         *Context
}
type CodeClientCert struct {
	clientCertificate *cert.Certificate
	code              string
}

func (c *CodeClientCert) GetServerName() string {
	return c.clientCertificate.ServerName
}
func NewClientCert(context *Context, userModel *db.UserModel) *ClientCert {
	return &ClientCert{context: context, codeClientCerts: make([]*CodeClientCert, 0), userModel: userModel}
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
		c.context.GetLog().Debug("loadAllUser", zap.String("Username", user.Username), zap.String("Code", user.Code), zap.String("CertPath", user.CertPath))
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
func (c *ClientCert) GetClientCerts() []*CodeClientCert {
	return c.codeClientCerts
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
		c.codeClientCerts = append(c.codeClientCerts, &CodeClientCert{clientCertificate: cc, code: code})
		return nil
	}
	return errors.New("用户名错误")
}
