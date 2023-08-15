package web

import "testing"

func TestName(t *testing.T) {

	server := NewServer()
	server.Init()
	server.Start()

}
