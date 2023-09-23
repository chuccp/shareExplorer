package traversal

import (
	"github.com/chuccp/shareExplorer/core"
	user2 "github.com/chuccp/shareExplorer/entity"
	"log"
	"time"
)

type Client struct {
	context *core.Context
}

func newClient(context *core.Context) *Client {
	return &Client{context: context}
}

func (c *Client) start() {
	time.Sleep(2 * time.Second)
	log.Println("!!!========================")
	address := c.context.GetConfigArray("traversal", "remote.address")
	go func() {
		for {
			time.Sleep(10 * time.Second)
			for _, addr := range address {
				c.register(addr)
				time.Sleep(1 * time.Second)
			}
		}
	}()
}
func (c *Client) register(address string) {

	var user user2.RemoteHost
	user.Username = "121212112"

	log.Println("777========================")
	client, err := c.context.GetHttpClient(address)
	if err != nil {
		return
	}
	str, err := client.PostJson("/traversal/register", &user)
	if err != nil {
		return
	}
	log.Println(str)
}
