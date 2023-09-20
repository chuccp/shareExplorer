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
	host := c.context.GetConfig("traversal", "remote.host")
	port := c.context.GetConfig("traversal", "remote.port")
	address := host + ":" + port
	go func() {
		for {
			c.register(address)
			time.Sleep(10 * time.Second)
		}
	}()
}
func (c *Client) register(address string) {

	var user user2.User
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
