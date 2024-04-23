package discover

import "testing"

func Test_mmm(t *testing.T) {
	var nodeMap map[ID]*mNode = make(map[ID]*mNode)

	id, err := StringToId("96e79218965eb72c92a549dd5a33011296e79218965eb72c92a549dd5a330112")
	if err != nil {
		t.Error(err)
	}
	value, ok := nodeMap[id]
	t.Log(value, ok)
	if !ok {
		nodeMap[id] = nil
	}

	id, err = StringToId("96e79218965eb72c92a549dd5a33011296e79218965eb72c92a549dd5a330112")
	if err != nil {
		t.Error(err)
	}
	value, ok = nodeMap[id]
	t.Log(value, ok)
}
