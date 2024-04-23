//go:build windows

package io

import (
	"fmt"
	"path/filepath"
	"syscall"
)

func IsHiddenFile(filename string) (bool, error) {
	basename := filepath.Base(filename)
	if basename[0:1] == "." {
		return true, nil
	}
	pointer, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		fmt.Errorf("IsHiddenFile %s", err)
		return false, err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		fmt.Errorf("IsHiddenFile %s", err)
		return false, err
	}
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
