package web

import "testing"

func TestWeb(t *testing.T) {

	var json = `{ "code":200,"data":"111111" }`

	response, err := JsonToResponse[string](json)
	if err != nil {
		t.Error(err)
	} else {
		println(response.Data)
	}

}
