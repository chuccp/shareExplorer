package entity

import "net"

const (
	SearchInit     = 2
	Searching      = 3
	SearchFailed   = 4
	SearchComplete = 5
)

type NodeStatus struct {
	address *net.UDPAddr
	Status  int    `json:"status"`
	Error   string `json:"error"`
	err     error
}

func NewNodeStatus() *NodeStatus {
	return &NodeStatus{Status: SearchInit}
}
func (s *NodeStatus) IsComplete() bool {
	return s.Status == SearchComplete
}
func (s *NodeStatus) IsOK() bool {
	return s.Status == SearchComplete && s.err == nil
}
func (s *NodeStatus) SearchFail(err error) {
	if err != nil {
		s.Error = err.Error()
		s.err = err
		s.Status = SearchFailed
	}
}
func (s *NodeStatus) StartSearch() {
	s.Status = Searching
}
func (s *NodeStatus) IsSearching() bool {
	return s.Status == Searching
}
func (s *NodeStatus) GetError() error {
	return s.err
}
func (s *NodeStatus) GetMsg() string {
	return "查找节点并登录中..."
}
func (s *NodeStatus) GetCode() int {

	if s.Status == Searching || s.Status == SearchInit {
		return 203
	}
	if s.Status == SearchFailed {
		return 204
	}
	return 200
}
func (s *NodeStatus) GetAddress() *net.UDPAddr {
	return s.address
}
func (s *NodeStatus) GetRemoteAddress() string {
	if s.address != nil {
		return s.address.String()
	}
	return ""
}
func (s *NodeStatus) SearchComplete(address *net.UDPAddr) {
	s.address = address
	s.Status = SearchComplete
	s.err = nil
	s.Error = ""
}
