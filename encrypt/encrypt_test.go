package encrypt

import (
	 "github.com/chuccp/utils/log"
	"testing"
)

func TestNewEncrypt(t *testing.T) {

	encrypt:=NewEncrypt("key")

	log.Info(encrypt.GenerateTLSConfig())

}