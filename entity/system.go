package entity

type System struct {
	HasInit       bool     `json:"hasInit"`
	RemoteAddress []string `json:"remoteAddress"`
}
