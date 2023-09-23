package traversal

import (
	"container/list"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/util"
	"github.com/chuccp/shareExplorer/web"
)

type Store struct {
	userList *list.List
	hostMap  map[string]*entity.RemoteHost
}

func (s *Store) AddUser(user *entity.RemoteHost) {
	username := user.Username
	ur, ok := s.hostMap[username]
	if !ok {
		s.hostMap[username] = user
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

func (s *Store) QueryPage(page *web.Page) []*entity.RemoteHost {
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
func (s *Store) Query(username string) (*entity.RemoteHost, bool) {
	u, ok := s.hostMap[username]
	return u, ok
}

func newStore() *Store {
	uList := list.New()
	return &Store{userList: uList, hostMap: make(map[string]*entity.RemoteHost)}
}
