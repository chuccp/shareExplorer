package util

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func MD5(data []byte) string {
	hash := md5.New()
	hash.Write(data)
	hashValue := hash.Sum(nil)
	md5Hash := hex.EncodeToString(hashValue)
	return md5Hash
}

func SplitPath(path string) []string {
	path = strings.ReplaceAll(path, "\\", "/")
	vs := strings.Split(path, "/")
	ps := make([]string, 0)
	for _, i2 := range vs {
		i2 = strings.TrimSpace(i2)
		if len(i2) == 0 {
			continue
		}
		ps = append(ps, i2)
	}
	return ps
}
