package entity

import "github.com/chuccp/shareExplorer/util"

type RemoteHost struct {
	ServerName   string `json:"serverName"`
	RemoteAddr   string `json:"remoteAddr"`
	CreateTime   string `json:"createTime"`
	UpdateTime   string `json:"updateTime"`
	LastLiveTime string `json:"lastLiveTime"`
	IsOnline     bool   `json:"isOnline"`
}

func NewRemoteHost(serverName string, remoteAddr string) *RemoteHost {
	u := &RemoteHost{ServerName: serverName, RemoteAddr: remoteAddr}
	u.CreateTime = util.NowTime()
	u.UpdateTime = util.NowTime()
	u.LastLiveTime = util.NowTime()
	u.IsOnline = true
	return u
}
