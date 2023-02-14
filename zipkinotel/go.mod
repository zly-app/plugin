module github.com/zly-app/plugin/zipkinotel

go 1.16

require (
	github.com/spf13/cast v1.3.0
	github.com/zly-app/zapp v1.1.16
	go.opentelemetry.io/otel v1.13.0
	go.opentelemetry.io/otel/exporters/jaeger v1.13.0
	go.opentelemetry.io/otel/exporters/zipkin v1.13.0
	go.opentelemetry.io/otel/sdk v1.13.0
	go.uber.org/zap v1.16.0
)
