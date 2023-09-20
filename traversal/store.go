package traversal

import (
	"container/list"
	"github.com/chuccp/shareExplorer/entity"
	"github.com/chuccp/shareExplorer/util"
	"github.com/chuccp/shareExplorer/web"
)

type Store struct {
	userList *list.List
	userMap  map[string]*entity.User
}

func (s *Store) AddUser(user *entity.User) {
	username := user.Username
	ur, ok := s.userMap[username]
	if !ok {
		s.userMap[username] = user
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

func (s *Store) QueryPage(page *web.Page) []*entity.User {
	us := make([]*entity.User, 0)
	index := 0
	start := page.PageNo * page.PageSize
	end := start + page.PageSize
	for ele := s.userList.Front(); ele != nil; ele = ele.Next() {
		index++
		if index >= start && index < end {
			us = append(us, (ele.Value).(*entity.User))
		}
	}
	return us
}
func (s *Store) Query(username string) (*entity.User, bool) {
	u, ok := s.userMap[username]
	return u, ok
}

func newStore() *Store {
	uList := list.New()
	return &Store{userList: uList, userMap: make(map[string]*entity.User)}
}
