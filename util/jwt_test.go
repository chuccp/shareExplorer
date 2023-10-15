package util

import (
	"testing"
)

func TestJwt(t *testing.T) {

	j := NewJwt()

	signedString, err := j.SignedSub("xxxxx")

	println(signedString, err)

	claims, err := j.ParseWithSub(signedString)

	println(claims, err)
	//ZXlKaGJHY2lPaUpJVXpJMU5pSXNJblI1Y0NJNklrcFhWQ0o5LmV5SkZlSEJwY21WelFYUWlPakUyT1Rjek5EazFPVGdzSWxWelpYSkpaQ0k2SWpFeU16UTBOU0o5LkFxQ29ZRGpPaXpoOEFiMkRsUUZSQWFiNjh3cENxbUlBdTFXRVhkWjRiYjQ
	//ZXlKaGJHY2lPaUpJVXpJMU5pSXNJblI1Y0NJNklrcFhWQ0o5LmV5SkZlSEJwY21WelFYUWlPakUyT1Rjek5EazJNVFFzSWxWelpYSkpaQ0k2SWpFeU16UTBOU0o5LjlqZC0wU3ZGMXNIN1gweU51SXAwWjByU1ZrZDNzaFpHNHNhUXJCU3Vsd2M
}
