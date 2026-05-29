package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type Application struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Project        string    `json:"project"`
	RepoURL        string    `json:"repoUrl"`
	Path           string    `json:"path"`
	TargetRevision string    `json:"targetRevision"`
	Destination    string    `json:"destination"`
	Namespace      string    `json:"namespace"`
	SyncPolicy     string    `json:"syncPolicy"`
	Health         string    `json:"health"`
	SyncStatus     string    `json:"syncStatus"`
	LastSyncedAt   time.Time `json:"lastSyncedAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ApplicationRevision struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"applicationId"`
	Revision      string    `json:"revision"`
	Message       string    `json:"message"`
	CreatedAt     time.Time `json:"createdAt"`
}

type RepositoryCredential struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	SecretRef string    `json:"secretRef"`
	CreatedAt time.Time `json:"createdAt"`
}

type PodLogRequest struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
	TailLines int64  `json:"tailLines"`
}
