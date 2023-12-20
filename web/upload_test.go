package web

import "testing"

func TestName(t *testing.T) {

	tempUploads, err := SplitFile("C:\\Users\\cooge\\Downloads\\apache-tomcat-8.5.97-windows-x64.zip", "C:\\Users\\cooge\\Downloads\\apache-tomcat-8.5.97-windows-x64(2).zip", 10)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(tempUploads)

	for _, v := range tempUploads {
		err := v.SaveUploaded()
		if err != nil {
			t.Log(err)
			return
		}
	}

}
