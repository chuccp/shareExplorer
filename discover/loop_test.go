package discover

import (
	"testing"
	"time"
)

func TestLoop(t *testing.T) {

	revalidate := time.NewTimer(time.Second * 5)
loop:
	for {
		select {
		case <-revalidate.C:
			t.Log("222222")
			break loop

		}
	}

	t.Log("11111111")

}
