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
	status  int
	err     error
}

func NewNodeStatus() *NodeStatus {
	return &NodeStatus{status: SearchInit}
}
func (s *NodeStatus) IsComplete() bool {
	return s.status == SearchComplete
}
func (s *NodeStatus) SearchFail(err error) {
	s.err = err
}
func (s *NodeStatus) GetError() error {
	return s.err
}
func (s *NodeStatus) GetAddress() *net.UDPAddr {
	return s.address
}
func (s *NodeStatus) SearchComplete(address *net.UDPAddr) {
	s.address = address
	s.status = SearchComplete
}
