package file

import (
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/webdav"
	"io"
	"log"
	"net/http"
	"testing"
)

func newDav() *webdav.Handler {
	webdav := &webdav.Handler{
		FileSystem: webdav.Dir(""),
		LockSystem: webdav.NewMemLS(),
		Prefix:     "/dav",
	}
	return webdav
}

var ddd = newDav()

func davHander(response http.ResponseWriter, request *http.Request) {

	v, _ := json.Marshal(request.Header)

	log.Println(request.RequestURI, request.Method, string(v))
	data, _ := io.ReadAll(request.Body)
	log.Println(string(data))

	ddd.ServeHTTP(response, request)
	v2, _ := json.Marshal(response.Header())
	log.Println(string(v2))
}

func TestName(t *testing.T) {
	var addr *string
	var path *string
	addr = flag.String("addr", ":8080", "") // listen端口，默认8080
	path = flag.String("path", ".", "")     // 文件路径，默认当前目录
	flag.Parse()
	fmt.Println("addr=", *addr, ", path=", *path) // 在控制台输出配置
	http.HandleFunc("/dav/", davHander)
	http.ListenAndServe(*addr, nil)
}
