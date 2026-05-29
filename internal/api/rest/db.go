package rest

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kubefusion/kubefusion/internal/store/postgres"
)

func (s *Server) dbStore() *postgres.Store {
	if s.DB == nil {
		return nil
	}
	return postgres.NewStore((*pgxpool.Pool)(s.DB))
}
