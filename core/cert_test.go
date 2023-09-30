package core

import "testing"

func TestParseGroupKey(t *testing.T) {

	ca, tls, err := ReadClientGroupPem("share.group.key")
	if err != nil {
		return
	}
	t.Log(ca, tls)
}
