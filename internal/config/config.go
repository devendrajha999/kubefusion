package config

import (
	"os"
)

type Config struct {
	HTTPAddr      string
	GRPCAddr      string
	PostgresDSN   string
	RedisAddr     string
	RedisPassword string
	JWTSecret     string
	KubeConfig    string
	OTLPEndpoint  string
	PrometheusURL string
}

func Load() Config {
	return Config{
		HTTPAddr:      get("HTTP_ADDR", ":8080"),
		GRPCAddr:      get("GRPC_ADDR", ":9090"),
		PostgresDSN:   get("POSTGRES_DSN", "postgres://kubefusion:kubefusion@postgres:5432/kubefusion?sslmode=disable"),
		RedisAddr:     get("REDIS_ADDR", "redis:6379"),
		RedisPassword: get("REDIS_PASSWORD", ""),
		JWTSecret:     get("JWT_SECRET", "change-me"),
		KubeConfig:    get("KUBECONFIG", ""),
		OTLPEndpoint:  get("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"),
		PrometheusURL: get("PROMETHEUS_URL", "http://prometheus-operated.monitoring.svc.cluster.local:9090"),
	}
}

func get(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}
