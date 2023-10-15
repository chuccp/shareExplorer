package util

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"strconv"
	"time"
)

type Jwt struct {
	key []byte
}

func (j *Jwt) SignedString(claims jwt.Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := t.SignedString(j.key)
	if err != nil {
		return "", err
	}
	return t.EncodeSegment([]byte(signedString)), err
}
func (j *Jwt) SignedSub(sub string) (string, error) {
	claim := jwt.MapClaims{"sub": sub, "exp": jwt.NewNumericDate(time.Now().Add(24 * time.Hour))}
	return j.SignedString(claim)
}
func (j *Jwt) ParseWithClaims(tokenString string) (jwt.MapClaims, error) {
	p := jwt.NewParser()
	segment, err := p.DecodeSegment(tokenString)
	if err != nil {
		return nil, err
	}
	token, err := p.ParseWithClaims(string(segment), jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.key, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}

func (j *Jwt) ParseWithSub(tokenString string) (string, error) {
	claims, err := j.ParseWithClaims(tokenString)
	if err != nil {
		return "", err
	}
	return claims.GetSubject()
}

func NewJwt() *Jwt {
	newUUID, _ := uuid.NewUUID()
	key := strconv.Itoa(int(time.Now().UnixMicro()))
	return &Jwt{key: []byte(newUUID.String() + key)}
}
