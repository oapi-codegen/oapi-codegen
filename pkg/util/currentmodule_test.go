package util

import "testing"

func TestCurrentModule(t *testing.T) {
	m, err := GetCurrentModule()
	if err != nil {
		t.Error(err)
	}
	println(m)
}
