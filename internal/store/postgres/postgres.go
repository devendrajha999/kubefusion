package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ DB *pgxpool.Pool }

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConns = 20
	return pgxpool.NewWithConfig(ctx, cfg)
}

func NewStore(db *pgxpool.Pool) *Store { return &Store{DB: db} }

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.DB.Exec(ctx, `
CREATE TABLE IF NOT EXISTS applications (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  project TEXT NOT NULL,
  repo_url TEXT NOT NULL,
  path TEXT NOT NULL,
  target_revision TEXT NOT NULL,
  destination TEXT NOT NULL,
  namespace TEXT NOT NULL,
  sync_policy TEXT NOT NULL,
  health TEXT NOT NULL,
  sync_status TEXT NOT NULL,
  last_synced_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE IF NOT EXISTS app_revisions (
  id TEXT PRIMARY KEY,
  application_id TEXT NOT NULL,
  revision TEXT NOT NULL,
  message TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE IF NOT EXISTS audit_events (
  id TEXT PRIMARY KEY,
  actor TEXT NOT NULL,
  action TEXT NOT NULL,
  target TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);`)
	return err
}
