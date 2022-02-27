package encrypt

import (
	log "github.com/chuccp/coke-log"
	"testing"
)

func TestNewEncrypt(t *testing.T) {

	encrypt:=NewEncrypt("key")

	log.Info(encrypt.GenerateTLSConfig())

}