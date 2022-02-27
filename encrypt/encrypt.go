package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"shareExplorer/file"
)

type Encrypt struct {
	localPath string
}
type Rsa struct {
	keyPEM  []byte
	certPEM []byte
}

func (encrypt *Encrypt) GenerateKey() ([]byte, []byte) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return certPEM, keyPEM
}
func (encrypt *Encrypt) GenerateNewTLSConfig() (*tls.Config, error) {
	fi, err := file.NewFile(encrypt.localPath)
	if err != nil {
		return nil, err
	}
	err = fi.MkDirs()
	if err != nil {
		return nil, err
	}
	keyFile, err1 := fi.Child("key.pem")
	certFile, err2 := fi.Child("cert.pem")
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	certPEM, keyPEM := encrypt.GenerateKey()
	if err != nil {
		return nil, err
	}
	err1 = keyFile.WriteBytes(keyPEM)
	err2 = certFile.WriteBytes(certPEM)
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-shareExplorer"},
	}, nil

}

func (encrypt *Encrypt) GenerateTLSConfig() (*tls.Config, error) {
	fi, err := file.NewFile(encrypt.localPath)
	if err != nil {
		return nil, err
	}
	keyFile, err1 := fi.Child("key.pem")
	certFile, err2 := fi.Child("cert.pem")
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	keyByte, has1, err1 := keyFile.Read()
	certByte, has2, err2 := certFile.Read()
	if has1 && has2 && err1 == nil && err2 == nil {
		tlsCert, err := tls.X509KeyPair(certByte, keyByte)
		if err != nil {
			panic(err)
		}
		return &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			NextProtos:   []string{"quic-shareExplorer"},
		}, nil
	}
	return encrypt.GenerateNewTLSConfig()
}

func NewEncrypt(localPath string) *Encrypt {
	return &Encrypt{localPath: localPath}
}
