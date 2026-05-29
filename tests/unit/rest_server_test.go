package rest

import "testing"

func TestNewServer(t *testing.T) {
	s := New("secret")
	if s == nil {
		t.Fatal("expected server")
	}
}
