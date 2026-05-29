package postgres

import (
	"context"
	"time"

	"github.com/kubefusion/kubefusion/internal/models"
)

func (s *Store) ListApplications(ctx context.Context) ([]models.Application, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,name,project,repo_url,path,target_revision,destination,namespace,sync_policy,health,sync_status,last_synced_at,created_at,updated_at FROM applications ORDER BY created_at DESC`)
	if err != nil { return nil, err }
	defer rows.Close()
	out := []models.Application{}
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID,&a.Name,&a.Project,&a.RepoURL,&a.Path,&a.TargetRevision,&a.Destination,&a.Namespace,&a.SyncPolicy,&a.Health,&a.SyncStatus,&a.LastSyncedAt,&a.CreatedAt,&a.UpdatedAt); err != nil { return nil, err }
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) ListRevisions(ctx context.Context, appID string) ([]models.ApplicationRevision, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,application_id,revision,message,created_at FROM app_revisions WHERE application_id=$1 ORDER BY created_at DESC`, appID)
	if err != nil { return nil, err }
	defer rows.Close()
	out := []models.ApplicationRevision{}
	for rows.Next() {
		var r models.ApplicationRevision
		if err := rows.Scan(&r.ID,&r.ApplicationID,&r.Revision,&r.Message,&r.CreatedAt); err != nil { return nil, err }
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) ListAudit(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,actor,action,target,created_at FROM audit_events ORDER BY created_at DESC LIMIT 1000`)
	if err != nil { return nil, err }
	defer rows.Close()
	out := []map[string]interface{}{}
	for rows.Next() {
		var id, actor, action, target string
		var created time.Time
		if err := rows.Scan(&id,&actor,&action,&target,&created); err != nil { return nil, err }
		out = append(out, map[string]interface{}{"id":id,"actor":actor,"action":action,"target":target,"createdAt":created})
	}
	return out, rows.Err()
}
