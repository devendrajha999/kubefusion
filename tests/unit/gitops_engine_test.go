package unit

import (
	"testing"

	"github.com/kubefusion/kubefusion/internal/gitops"
	"github.com/kubefusion/kubefusion/internal/models"
)

func TestDetectDrift(t *testing.T) {
	e := gitops.NewEngine()
	app := models.Application{TargetRevision: "main", SyncStatus: "OutOfSync"}
	status, _ := e.DetectDrift(app)
	if status != "Drifted" {
		t.Fatalf("expected Drifted, got %s", status)
	}
}
