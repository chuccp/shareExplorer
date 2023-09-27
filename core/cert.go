package core

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/chuccp/kuic/cert"
	"github.com/chuccp/kuic/util"
	"github.com/chuccp/shareExplorer/io"
)

type Cert struct {
	CaPath      string
	CertPath    string
	KeyPath     string
	certPool    *x509.CertPool
	certificate *tls.Certificate
}

func (c *Cert) getTlsConfig() *tls.Config {
	config := &tls.Config{
		ClientCAs:    c.certPool,
		ClientAuth:   tls.VerifyClientCertIfGiven,
		Certificates: []tls.Certificate{*c.certificate},
	}
	return config
}

func initCert(caPath string, certPath string, keyPath string) (*Cert, error) {
	var cert = &Cert{
		CaPath:   caPath,
		CertPath: certPath,
		KeyPath:  keyPath,
	}
	cert.certPool = x509.NewCertPool()
	err := CreateNotExistCertGroup(caPath, certPath, keyPath)
	if err != nil {
		return nil, err
	}
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	} else {
		cert.certificate = &certificate
		return cert, nil
	}
}
func InitServerCert() (*Cert, error) {
	caPath := "server.cer"
	keyPath := "server.key"
	certPath := "server.crt"
	return initCert(caPath, certPath, keyPath)
}
func InitClientCert(name string) (*Cert, error) {
	caPath := name + ".client.cer"
	keyPath := name + ".client.key"
	certPath := name + ".client.crt"
	return initCert(caPath, certPath, keyPath)
}

func GenerateClientGroupPem(serverCa, clientCert, clientKey string) ([]byte, error) {
	data := make([]byte, 0)
	ca, err := io.ReadFile(serverCa)
	if err != nil {
		return nil, err
	}
	data = append(data, ca...)
	data = append(data, '\n')

	cert, err := io.ReadFile(clientCert)
	if err != nil {
		return nil, err
	}
	data = append(data, cert...)
	data = append(data, '\n')

	key, err := io.ReadFile(clientKey)
	if err != nil {
		return nil, err
	}
	data = append(data, key...)
	return data, nil
}

func CreateNotExistCertGroup(caPath string, certPath string, keyPath string) (err error) {
	caIsExists, err := util.ExistsFile(caPath)
	if err != nil {
		return err
	}
	certIsExists, err := util.ExistsFile(certPath)
	if err != nil {
		return err
	}
	keyIsExists, err := util.ExistsFile(keyPath)
	if err != nil {
		return err
	}
	if caIsExists && certIsExists && keyIsExists {
		return nil
	}
	err = cert.CreateCertGroup(nil, caPath, certPath, keyPath)
	return err
}
