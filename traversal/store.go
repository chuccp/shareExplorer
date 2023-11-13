package traversal

import (
	"container/list"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/util"
	"github.com/chuccp/shareExplorer/web"
)

type ClientStore struct {
	userList *list.List
	hostMap  map[string]*entity.RemoteHost
}

func (s *ClientStore) AddUser(user *entity.RemoteHost) {
	serverName := user.ServerName
	ur, ok := s.hostMap[serverName]
	if !ok {
		s.hostMap[serverName] = user
		s.userList.PushBack(user)
	} else {
		if ur.RemoteAddr != user.RemoteAddr {
			ur.UpdateTime = util.NowTime()
			ur.RemoteAddr = user.RemoteAddr
		} else {
			ur.LastLiveTime = util.NowTime()
		}
	}
}

func (s *ClientStore) QueryPage(page *web.Page) []*entity.RemoteHost {
	us := make([]*entity.RemoteHost, 0)
	index := 0
	start := page.PageNo * page.PageSize
	end := start + page.PageSize
	for ele := s.userList.Front(); ele != nil; ele = ele.Next() {
		index++
		if index >= start && index < end {
			us = append(us, (ele.Value).(*entity.RemoteHost))
		}
	}
	return us
}
func (s *ClientStore) Query(username string) (*entity.RemoteHost, bool) {
	u, ok := s.hostMap[username]
	return u, ok
}

func newClientStore() *ClientStore {
	uList := list.New()
	return &ClientStore{userList: uList, hostMap: make(map[string]*entity.RemoteHost)}
}
