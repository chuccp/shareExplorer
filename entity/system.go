package entity

type System struct {
	HasInit       bool     `json:"hasInit"`
	HasSignIn     bool     `json:"hasSignIn"`
	RemoteAddress []string `json:"remoteAddress"`
}
