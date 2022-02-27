package quic

import (
	"bufio"
	"github.com/chuccp/cokePush/config"
	"github.com/lucas-clemente/quic-go"
	"golang.org/x/net/context"
	"log"
	"shareExplorer/encrypt"
	"strconv"
)

type Server struct {
	port int
	keyPath string
}

func (server *Server) Start() error {

	encrypt := encrypt.NewEncrypt(server.keyPath)
	tls, err := encrypt.GenerateTLSConfig()
	if err != nil {
		return err
	}
	listen, err := quic.ListenAddr(":"+strconv.Itoa(server.port), tls, nil)
	if err != nil {
		panic(err)
	}
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
	return err

}
func (server *Server) Init(cfg *config.Config) {
	server.port = cfg.GetInt("share.port")
	server.keyPath = cfg.GetString("share.keyPath")
}
