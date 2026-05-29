package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kubefusion/kubefusion/internal/audit"
	"github.com/kubefusion/kubefusion/internal/gitops"
	"github.com/kubefusion/kubefusion/internal/kube"
	"github.com/kubefusion/kubefusion/internal/models"
)

type Server struct {
	JWTSecret      string
	Apps           []models.Application
	Revisions      map[string][]models.ApplicationRevision
	RepoCreds      []models.RepositoryCredential
	SyncWindowOpen bool
	GitOps         *gitops.Engine
	Audit          *audit.Store
	Kube           *kube.Client
	DB             *pgxpool.Pool
}

func New(jwtSecret string, kubeClient *kube.Client, db *pgxpool.Pool) *Server {
	return &Server{JWTSecret: jwtSecret, Apps: []models.Application{}, Revisions: map[string][]models.ApplicationRevision{}, RepoCreds: []models.RepositoryCredential{}, SyncWindowOpen: true, GitOps: gitops.NewEngine(), Audit: audit.NewStore(), Kube: kubeClient, DB: db}
}

func (s *Server) Register(r *gin.Engine) { /* unchanged */
	r.POST("/api/v1/auth/login", s.login)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	a := r.Group("/api/v1")
	a.Use(requireAuth(s.JWTSecret))
	{
		a.GET("/applications", s.listApplications)
		a.POST("/applications", requireRole("admin"), s.createApplication)
		a.POST("/applications/:id/sync", requireRole("admin", "operator"), s.syncApplication)
		a.GET("/applications/:id/drift", s.detectDrift)
		a.GET("/applications/:id/history", s.applicationHistory)
		a.POST("/applications/:id/rollback", requireRole("admin", "operator"), s.rollbackApplication)
		a.GET("/repositories/credentials", requireRole("admin"), s.listRepositoryCredentials)
		a.POST("/repositories/credentials", requireRole("admin"), s.createRepositoryCredential)
		a.GET("/clusters", s.listClusters)
		a.GET("/clusters/:name/nodes", s.listNodes)
		a.GET("/clusters/:name/pods", s.listPods)
		a.POST("/clusters/:name/pods/logs", s.getPodLogs)
		a.GET("/clusters/:name/pods/logs/stream", s.streamPodLogs)
		a.POST("/clusters/:name/pods/exec", requireRole("admin", "operator"), s.execPod)
		a.GET("/sync-windows", s.getSyncWindow)
		a.POST("/sync-windows/toggle", requireRole("admin"), s.toggleSyncWindow)
		a.GET("/audit/events", requireRole("admin"), s.listAudit)
	}
}

func (s *Server) audit(c *gin.Context, action, target string) {
	actor := c.GetString("username")
	if actor == "" { actor = "anonymous" }
	e := audit.Event{ID: uuid.NewString(), Actor: actor, Action: action, Target: target, CreatedAt: time.Now().UTC()}
	s.Audit.Add(e)
	if s.DB != nil { _, _ = s.DB.Exec(context.Background(), "INSERT INTO audit_events (id, actor, action, target, created_at) VALUES ($1,$2,$3,$4,$5)", e.ID, e.Actor, e.Action, e.Target, e.CreatedAt) }
}

func (s *Server) listApplications(c *gin.Context) {
	if ds := s.dbStore(); ds != nil {
		apps, err := ds.ListApplications(c.Request.Context())
		if err == nil { c.JSON(http.StatusOK, apps); return }
	}
	c.JSON(http.StatusOK, s.Apps)
}

func (s *Server) createApplication(c *gin.Context) {
	var app models.Application
	if err := c.BindJSON(&app); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	now := time.Now().UTC(); app.ID = uuid.NewString(); app.CreatedAt = now; app.UpdatedAt = now; app.Health = "Healthy"; app.SyncStatus = "OutOfSync"; app.LastSyncedAt = now
	s.Apps = append(s.Apps, app)
	if s.DB != nil { _, _ = s.DB.Exec(context.Background(), `INSERT INTO applications (id,name,project,repo_url,path,target_revision,destination,namespace,sync_policy,health,sync_status,last_synced_at,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, app.ID, app.Name, app.Project, app.RepoURL, app.Path, app.TargetRevision, app.Destination, app.Namespace, app.SyncPolicy, app.Health, app.SyncStatus, app.LastSyncedAt, app.CreatedAt, app.UpdatedAt) }
	s.audit(c, "application.create", app.Name); c.JSON(http.StatusCreated, app)
}

func (s *Server) syncApplication(c *gin.Context) { if !s.SyncWindowOpen { c.JSON(http.StatusConflict, gin.H{"error":"sync window closed"}); return }; id := c.Param("id"); for i := range s.Apps { if s.Apps[i].ID==id { rev:=s.GitOps.Sync(&s.Apps[i], true); s.Revisions[id]=append(s.Revisions[id], rev); if s.DB!=nil { _, _ = s.DB.Exec(context.Background(),"INSERT INTO app_revisions (id, application_id, revision, message, created_at) VALUES ($1,$2,$3,$4,$5)",rev.ID,rev.ApplicationID,rev.Revision,rev.Message,rev.CreatedAt) }; s.audit(c,"application.sync",s.Apps[i].Name); c.JSON(http.StatusOK,s.Apps[i]); return } }; c.JSON(http.StatusNotFound, gin.H{"error":"application not found"}) }
func (s *Server) detectDrift(c *gin.Context) { id:=c.Param("id"); for i := range s.Apps { if s.Apps[i].ID==id { status,diffs:=s.GitOps.DetectDrift(s.Apps[i]); c.JSON(http.StatusOK, gin.H{"status":status,"diffs":diffs}); return } }; c.JSON(http.StatusNotFound, gin.H{"error":"application not found"}) }
func (s *Server) applicationHistory(c *gin.Context) {
	id := c.Param("id")
	if ds := s.dbStore(); ds != nil {
		h, err := ds.ListRevisions(c.Request.Context(), id)
		if err == nil { c.JSON(http.StatusOK, h); return }
	}
	c.JSON(http.StatusOK, s.Revisions[id])
}
func (s *Server) rollbackApplication(c *gin.Context) { id:=c.Param("id"); var req struct{ Revision string `json:"revision"` }; if err:=c.BindJSON(&req); err!=nil || req.Revision=="" { c.JSON(http.StatusBadRequest, gin.H{"error":"revision required"}); return }; for i:=range s.Apps { if s.Apps[i].ID==id { rev:=s.GitOps.Rollback(&s.Apps[i], req.Revision); s.Revisions[id]=append(s.Revisions[id], rev); if s.DB!=nil { _, _ = s.DB.Exec(context.Background(),"INSERT INTO app_revisions (id, application_id, revision, message, created_at) VALUES ($1,$2,$3,$4,$5)",rev.ID,rev.ApplicationID,rev.Revision,rev.Message,rev.CreatedAt) }; s.audit(c,"application.rollback",fmt.Sprintf("%s:%s",s.Apps[i].Name,req.Revision)); c.JSON(http.StatusOK,s.Apps[i]); return } }; c.JSON(http.StatusNotFound, gin.H{"error":"application not found"}) }
func (s *Server) listRepositoryCredentials(c *gin.Context) { c.JSON(http.StatusOK, s.RepoCreds) }
func (s *Server) createRepositoryCredential(c *gin.Context) { var cred models.RepositoryCredential; if err:=c.BindJSON(&cred); err!=nil { c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()}); return }; cred.ID=uuid.NewString(); cred.CreatedAt=time.Now().UTC(); s.RepoCreds=append(s.RepoCreds,cred); s.audit(c,"repository.credential.create",cred.Name); c.JSON(http.StatusCreated,cred) }
func (s *Server) listClusters(c *gin.Context) { c.JSON(http.StatusOK, []gin.H{{"name":"in-cluster","status":"Healthy","server":"https://kubernetes.default.svc"}}) }
func (s *Server) listNodes(c *gin.Context) { if s.Kube==nil { c.JSON(http.StatusOK, []gin.H{{"name":"node-a","status":"True"}}); return }; nodes,err:=s.Kube.ListNodes(c.Request.Context()); if err!=nil { c.JSON(http.StatusBadGateway, gin.H{"error":err.Error()}); return }; c.JSON(http.StatusOK,nodes) }
func (s *Server) listPods(c *gin.Context) { ns:=c.Query("namespace"); if s.Kube==nil { c.JSON(http.StatusOK, []gin.H{{"namespace":"default","name":"nginx","status":"Running"}}); return }; pods,err:=s.Kube.ListPods(c.Request.Context(),ns); if err!=nil { c.JSON(http.StatusBadGateway, gin.H{"error":err.Error()}); return }; c.JSON(http.StatusOK,pods) }
func (s *Server) getPodLogs(c *gin.Context) { var req models.PodLogRequest; if err:=c.BindJSON(&req); err!=nil { c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()}); return }; if req.Namespace==""||req.Pod=="" { c.JSON(http.StatusBadRequest, gin.H{"error":"namespace and pod are required"}); return }; if req.TailLines==0 { req.TailLines=200 }; if s.Kube==nil { c.JSON(http.StatusOK, gin.H{"stream":false,"lines":[]string{"log line 1"},"target":req.Pod}); return }; lines,err:=s.Kube.PodLogs(c.Request.Context(),req.Namespace,req.Pod,req.Container,req.TailLines); if err!=nil { c.JSON(http.StatusBadGateway, gin.H{"error":err.Error()}); return }; c.JSON(http.StatusOK, gin.H{"stream":false,"lines":lines,"target":req.Pod}) }
func (s *Server) streamPodLogs(c *gin.Context) { ns:=c.Query("namespace"); pod:=c.Query("pod"); container:=c.Query("container"); if ns==""||pod=="" { c.JSON(http.StatusBadRequest, gin.H{"error":"namespace and pod are required"}); return }; if s.Kube==nil { c.JSON(http.StatusNotImplemented, gin.H{"error":"kubernetes client unavailable"}); return }; r,err:=s.Kube.StreamPodLogs(c.Request.Context(),ns,pod,container,200); if err!=nil { c.JSON(http.StatusBadGateway, gin.H{"error":err.Error()}); return }; defer r.Close(); c.Writer.Header().Set("Content-Type","text/event-stream"); c.Writer.Header().Set("Cache-Control","no-cache"); c.Writer.Header().Set("Connection","keep-alive"); c.Status(http.StatusOK); flusher,ok:=c.Writer.(http.Flusher); if !ok { c.JSON(http.StatusInternalServerError, gin.H{"error":"stream unsupported"}); return }; _=kube.ScanLines(r, func(line string) error { _,e:=c.Writer.Write([]byte("data: "+line+"\n\n")); if e!=nil { return e }; flusher.Flush(); return nil }) }
func (s *Server) execPod(c *gin.Context) { var req struct { Namespace string `json:"namespace"`; Pod string `json:"pod"`; Container string `json:"container"`; Command []string `json:"command"` }; if err:=c.BindJSON(&req); err!=nil { c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()}); return }; if req.Namespace==""||req.Pod==""||len(req.Command)==0 { c.JSON(http.StatusBadRequest, gin.H{"error":"namespace, pod, and command are required"}); return }; if s.Kube==nil { c.JSON(http.StatusNotImplemented, gin.H{"error":"kubernetes client unavailable"}); return }; stdout,stderr,err:=s.Kube.ExecOnce(c.Request.Context(),req.Namespace,req.Pod,req.Container,req.Command); if err!=nil { c.JSON(http.StatusBadGateway, gin.H{"error":err.Error(),"stdout":stdout,"stderr":stderr}); return }; c.JSON(http.StatusOK, gin.H{"stdout":stdout,"stderr":stderr}) }
func (s *Server) getSyncWindow(c *gin.Context) { state:="open"; if !s.SyncWindowOpen { state="closed" }; c.JSON(http.StatusOK, gin.H{"state":state}) }
func (s *Server) toggleSyncWindow(c *gin.Context) { s.SyncWindowOpen=!s.SyncWindowOpen; s.audit(c,"sync-window.toggle",fmt.Sprintf("open=%t",s.SyncWindowOpen)); s.getSyncWindow(c) }
func (s *Server) listAudit(c *gin.Context) {
	if ds := s.dbStore(); ds != nil {
		e, err := ds.ListAudit(c.Request.Context())
		if err == nil { c.JSON(http.StatusOK, e); return }
	}
	c.JSON(http.StatusOK, s.Audit.List())
}
