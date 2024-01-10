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
func (s *NodeStatus) SearchFail(err error) {
	s.Error = err.Error()
	s.err = err
}
func (s *NodeStatus) GetError() error {
	return s.err
}
func (s *NodeStatus) GetMsg() string {
	return "正在查找"
}
func (s *NodeStatus) GetAddress() *net.UDPAddr {
	return s.address
}
func (s *NodeStatus) SearchComplete(address *net.UDPAddr) {
	s.address = address
	s.Status = SearchComplete
	s.err = nil
	s.Error = ""
}
