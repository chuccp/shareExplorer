package util

import "testing"

func TestName(t *testing.T) {

	t.Log(ReplaceAllRegex("/*name/aaa", "\\*[a-zA-Z]+", ".[a-zA-Z]+"))

}
func TestName222(t *testing.T) {

	t.Log(IsMatchPath("/aaa/aaa", "/aaa/aaa"))

}
