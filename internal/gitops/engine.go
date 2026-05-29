package gitops

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kubefusion/kubefusion/internal/models"
)

type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) DetectDrift(app models.Application) (string, []string) {
	if strings.Contains(strings.ToLower(app.TargetRevision), "head") {
		return "Unknown", []string{"Unable to guarantee immutability when target revision is HEAD"}
	}
	if app.SyncStatus == "Synced" {
		return "InSync", nil
	}
	return "Drifted", []string{"Live manifest differs from desired state for one or more resources"}
}

func (e *Engine) Sync(app *models.Application, manual bool) models.ApplicationRevision {
	app.SyncStatus = "Synced"
	app.LastSyncedAt = time.Now().UTC()
	app.UpdatedAt = app.LastSyncedAt
	msg := "auto-sync"
	if manual {
		msg = "manual-sync"
	}
	return models.ApplicationRevision{
		ID:            uuid.NewString(),
		ApplicationID: app.ID,
		Revision:      app.TargetRevision,
		Message:       msg,
		CreatedAt:     time.Now().UTC(),
	}
}

func (e *Engine) Rollback(app *models.Application, toRevision string) models.ApplicationRevision {
	app.TargetRevision = toRevision
	app.SyncStatus = "OutOfSync"
	app.UpdatedAt = time.Now().UTC()
	return models.ApplicationRevision{
		ID:            uuid.NewString(),
		ApplicationID: app.ID,
		Revision:      toRevision,
		Message:       fmt.Sprintf("rollback-to-%s", toRevision),
		CreatedAt:     time.Now().UTC(),
	}
}
