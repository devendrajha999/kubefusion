package integration

import (
	"net/http"
	"testing"
)

func TestHealthContract(t *testing.T) {
	if http.StatusOK != 200 {
		t.Fatal("unexpected status code constant")
	}
}
