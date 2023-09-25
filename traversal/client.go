package traversal

import (
	"errors"
	"github.com/chuccp/kuic/http"
	"github.com/chuccp/shareExplorer/core"
	"github.com/chuccp/shareExplorer/web"
)

type Client struct {
	context       *core.Context
	remoteAddress string
}

func (c *Client) Register() error {

	return nil
}
func (c *Client) Connect() error {
	_, err := c.getString("/traversal/connect")
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) getString(url string) (string, error) {
	cl, err := c.getClient()
	if err != nil {
		return "", err
	}

	jsonString, err := cl.Get(url)
	if err != nil {
		return "", err
	}
	response, err := web.JsonToResponse[string](jsonString)
	if err != nil {
		return "", err
	}
	if response.IsOk() {
		return response.Data, nil
	}
	return "", errors.New(response.Data)
}

func (c *Client) getJsonValue(url string, value any) error {

	//cl, err := c.getClient()
	//if err != nil {
	//	return "", err
	//}
	//
	//_, err = cl.Get(url)
	//if err != nil {
	//	return "", err
	//}
	return nil
}

func (c *Client) getClient() (*http.Client, error) {
	client, err := c.context.GetHttpClient(c.remoteAddress)
	return client, err
}

func newClient(context *core.Context, remoteAddress string) *Client {
	return &Client{context: context, remoteAddress: remoteAddress}
}

//	func (c *Client) start() {
//		time.Sleep(2 * time.Second)
//		log.Println("!!!========================")
//		address := c.context.GetConfigArray("traversal", "remote.address")
//		go func() {
//			for {
//				time.Sleep(10 * time.Second)
//				for _, addr := range address {
//					c.register()
//					time.Sleep(1 * time.Second)
//				}
//			}
//		}()
//	}
//func (c *Client) register() {
//
//	//var user user2.RemoteHost
//	//user.Username = "121212112"
//	//
//	//log.Println("777========================")
//	//client, err := c.context.GetHttpClient(address)
//	//if err != nil {
//	//	return
//	//}
//	//str, err := client.PostJson("/traversal/register", &user)
//	//if err != nil {
//	//	return
//	//}
//	//log.Println(str)
//}
