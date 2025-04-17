package main

import "testing"

func TestDummyFail(t *testing.T) {
	t.Errorf("This is a dummy failing test")
}
