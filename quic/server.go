package quic

import (
	"github.com/lucas-clemente/quic-go"
	"golang.org/x/net/context"
	"shareExplorer/core"
	"shareExplorer/encrypt"
	"strconv"
)

type Server struct {
	port    int
	keyPath string
}

func (server *Server) Start() {
	encrypt := encrypt.NewEncrypt(server.keyPath)
	tls, err := encrypt.GenerateTLSConfig()
	if err != nil {
		panic(err)
	}
	listen, err := quic.ListenAddr(":"+strconv.Itoa(server.port), tls, nil)
	if err != nil {
		panic(err)
	}
	for {
		session, err2 := listen.Accept(context.Background())
		if err2 != nil {
			break
		}
		stream, err3 := session.OpenStream()
		if err3 == nil {
			conn := newConn(stream)
			go conn.Start()
		}
	}
}
func (server *Server) Init(ctx *core.Context) {
	server.port = ctx.GetConfig().GetInt("share.port")
	server.keyPath = ctx.GetConfig().GetString("share.keyPath")
}
