package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kubefusion/kubefusion/internal/api/rest"
	"github.com/kubefusion/kubefusion/internal/config"
	"github.com/kubefusion/kubefusion/internal/kube"
	"github.com/kubefusion/kubefusion/internal/store/postgres"
	"github.com/kubefusion/kubefusion/pkg/logger"
	"github.com/kubefusion/kubefusion/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	cfg := config.Load()
	log := logger.New()
	defer log.Sync()

	ctx := context.Background()
	shutdownOTel, err := telemetry.Init(ctx, cfg.OTLPEndpoint)
	if err == nil {
		defer shutdownOTel(ctx)
	}

	var kubeClient *kube.Client
	if kc, ke := kube.New(cfg.KubeConfig); ke == nil {
		kubeClient = kc
	} else {
		log.Sugar().Warnf("kube client disabled: %v", ke)
	}

	var db *pgxpool.Pool
	if pool, de := postgres.Connect(ctx, cfg.PostgresDSN); de == nil {
		db = pool
		defer db.Close()
		_ = postgres.NewStore(db).Migrate(ctx)
	} else {
		log.Sugar().Warnf("postgres disabled: %v", de)
	}

	r := gin.New()
	r.Use(gin.Recovery(), otelgin.Middleware("kubefusion-http"))
	rest.New(cfg.JWTSecret, kubeClient, db).Register(r)

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			log.Sugar().Fatalf("server failed: %v", e)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
