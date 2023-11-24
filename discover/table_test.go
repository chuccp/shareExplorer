package discover

import (
	"crypto/rand"
	"log"
	"net"
	"testing"
)

func TestRun(t *testing.T) {

	var id ID
	rand.Read(id[:])
	t.Log(id)
	t.Log(id.IsBlank())

}
func TestAdd(t *testing.T) {
	var id ID
	rand.Read(id[:])

	localNode, err := createLocalNode(id.String())
	if err != nil {
		log.Println(err)
	}

	table := NewTable(nil, localNode)

	var id2 ID
	rand.Read(id2[:])

	n, err := createLocalNode(id2.String())
	if err != nil {
		log.Println(err)
	}
	n.isNatClient = "true"
	n.isNatServer = "true"
	n.isServer = "true"
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1236")
	n.addr = addr
	table.addSeenNode(wrapNode(n))

	table.addNursery(addr)
	table.addNursery(addr)
}
func TestAdd2(t *testing.T) {
	var id ID
	rand.Read(id[:])

	localNode, err := createLocalNode(id.String())
	if err != nil {
		log.Println(err)
	}

	table := NewTable(nil, localNode)

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1236")

	table.addNursery(addr)
	//table.addNursery(addr)
	table.doRefresh()
}
