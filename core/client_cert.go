package core

import (
	"errors"
	"github.com/chuccp/kuic/cert"
	"github.com/chuccp/shareExplorer/db"
	"github.com/chuccp/shareExplorer/util"
	"go.uber.org/zap"
	"sync"
)

type ClientCert struct {
	codeClientCerts []*CodeClientCert
	userModel       *db.UserModel
	context         *Context
	rLock           *sync.RWMutex
}
type CodeClientCert struct {
	clientCertificate *cert.Certificate
	code              string
}

func (c *CodeClientCert) GetServerName() string {
	return c.clientCertificate.ServerName
}
func NewClientCert(context *Context, userModel *db.UserModel) *ClientCert {
	return &ClientCert{context: context, codeClientCerts: make([]*CodeClientCert, 0), userModel: userModel, rLock: new(sync.RWMutex)}
}
func (c *ClientCert) init() error {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	err := c._loadAllUser()
	if err != nil {
		return err
	}
	return nil
}
func (c *ClientCert) _loadAllUser() error {
	users, err := c.userModel.QueryAllClientUser()
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
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	user, err := c.userModel.QueryOneUser(username, code)
	if err != nil {
		return err
	}
	path := user.CertPath
	return c.LoadCert(username, code, path)
}

func (c *ClientCert) LoadAndDeleteUser(username string, code string, oldCode string) error {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	user, err := c.userModel.QueryOneUser(username, code)
	if err != nil {
		return err
	}
	path := user.CertPath
	err = c.LoadCert(username, code, path)
	if err == nil {
		c._deleteCert(oldCode)
	}
	return err
}

func (c *ClientCert) GetClientCerts() []*CodeClientCert {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	return c.codeClientCerts
}
func (c *ClientCert) _getCert(username string, code string) (*cert.Certificate, bool) {
	for _, certificate := range c.codeClientCerts {
		if certificate.clientCertificate.UserName == username && certificate.code == code {
			return certificate.clientCertificate, true
		}
	}
	return nil, false
}

func (c *ClientCert) GetCertByCode(username string, code string) (*cert.Certificate, bool) {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	return c._getCert(username, code)
}
func (c *ClientCert) LoadCert(username string, code string, path string) error {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
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
func (c *ClientCert) _deleteCert(code string) {
	for _, clientCert := range c.codeClientCerts {
		if code == clientCert.code {
			c.codeClientCerts = util.DeleteElement(c.codeClientCerts, clientCert)
			break
		}
	}
}
