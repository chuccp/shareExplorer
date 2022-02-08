package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/lucas-clemente/quic-go"
	"log"
	"math/big"
)
func generateTLSConfig() *tls.Config {
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

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}
func main_() {

	listen, err := quic.ListenAddr("0.0.0.0:4242", generateTLSConfig(), nil)
	if err == nil {
		log.Println("quick 启动=====")
		for {
			session, err2 := listen.Accept(context.Background())
			if err2 != nil {
				break
			}
			go func() {
				stream, err3 := session.OpenStream()
				if err3 == nil {
					read := bufio.NewReader(stream)
					for {
						line, _, err4 := read.ReadLine()
						if err4 == nil {
							log.Println(string(line))
						}
					}
				}
			}()
		}

	}else{
		log.Print(err)
	}
}
