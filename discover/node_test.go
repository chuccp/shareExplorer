package discover

import "testing"

func TestLogDist(t *testing.T) {

	id := wrapId([]byte{0, 1})

	t.Log(id)

	id2 := wrapId([]byte{0, 0})
	t.Log(id2)
	num := LogDist(id, id2)

	t.Log(num)

}
