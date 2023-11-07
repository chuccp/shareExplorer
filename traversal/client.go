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
	_, err := c.getRequestString("/traversal/connect")
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) ClientSignIn(username string, password string) error {
	_, err := c.getRequestString("/traversal/connect")
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) getRequestString(url string) (string, error) {
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

	return nil
}

func (c *Client) getClient() (*http.Client, error) {
	client, err := c.context.GetHttpClient(c.remoteAddress)
	return client, err
}

func newClient(context *core.Context, remoteAddress string) *Client {
	return &Client{context: context, remoteAddress: remoteAddress}
}
