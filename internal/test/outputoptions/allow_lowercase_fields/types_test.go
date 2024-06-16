package allowlowercasefields

import "testing"

func TestMyItemCompiles(t *testing.T) {
	_ = MyItem{
		Name: "a string",
		age:  1_000,
	}
}
