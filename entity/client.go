package entity

type Client struct {
	ServerName string `json:"serverName"`
	Username   string `json:"username"`
	UserId     int    `json:"id"`
	Code       string `json:"code"`
}
