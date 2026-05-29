module github.com/kubefusion/kubefusion

go 1.23

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0
	github.com/jackc/pgx/v5 v5.6.0
	github.com/prometheus/client_golang v1.20.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.53.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.66.0
	google.golang.org/protobuf v1.34.2
	k8s.io/apimachinery v0.31.0
	k8s.io/client-go v0.31.0
)
