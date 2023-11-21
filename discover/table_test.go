package discover

import (
	"crypto/rand"
	"testing"
)

func TestRun(t *testing.T) {

	var id ID
	rand.Read(id[:])
	t.Log(id)
	t.Log(id.IsBlank())

}
