package entity

type System struct {
	HasInit     bool `json:"hasInit"`
	HasSignIn   bool `json:"hasSignIn"`
	IsServer    bool `json:"isServer"`
	IsNatServer bool `json:"isNatServer"`
	IsClient    bool `json:"isClient"`

	RemoteAddress []string `json:"remoteAddress"`
}
