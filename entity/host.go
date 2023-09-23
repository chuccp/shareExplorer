package entity

import "github.com/chuccp/shareExplorer/util"

type RemoteHost struct {
	Username     string `json:"username"`
	RemoteAddr   string `json:"remoteAddr"`
	CreateTime   string `json:"createTime"`
	UpdateTime   string `json:"updateTime"`
	LastLiveTime string `json:"lastLiveTime"`
	IsOnline     bool   `json:"isOnline"`
}

func NewUser(username string, remoteAddr string) *RemoteHost {
	u := &RemoteHost{Username: username, RemoteAddr: remoteAddr}
	u.CreateTime = util.NowTime()
	u.UpdateTime = util.NowTime()
	u.LastLiveTime = util.NowTime()
	u.IsOnline = true
	return u
}
