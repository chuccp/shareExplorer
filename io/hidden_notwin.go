//go:build !windows

package io

import (
	"path/filepath"
)

func IsHiddenFile(filename string) (bool, error) {
	basename := filepath.Base(filename)
	if basename[0:1] == "." {
		return true, nil
	}
	return false, nil
}
