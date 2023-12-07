package entity

type System struct {
	HasInit       bool     `json:"hasInit"`
	HasSignIn     bool     `json:"hasSignIn"`
	IsServer      bool     `json:"isServer"`
	HasServer     bool     `json:"hasServer"`
	RemoteAddress []string `json:"remoteAddress"`
}
