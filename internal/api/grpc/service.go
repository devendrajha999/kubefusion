package grpc

import (
	"context"

	"github.com/kubefusion/kubefusion/internal/models"
)

type Service struct {
	Apps []models.Application
}

type ListApplicationsRequest struct{}
type Application struct {
	ID, Name, Project, SyncStatus, Health string
}
type ListApplicationsResponse struct{ Items []Application }

func (s *Service) ListApplications(ctx context.Context, req *ListApplicationsRequest) (*ListApplicationsResponse, error) {
	out := make([]Application, 0, len(s.Apps))
	for _, a := range s.Apps {
		out = append(out, Application{ID: a.ID, Name: a.Name, Project: a.Project, SyncStatus: a.SyncStatus, Health: a.Health})
	}
	return &ListApplicationsResponse{Items: out}, nil
}
