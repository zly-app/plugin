module github.com/zly-app/plugin/otlp

go 1.16

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/spf13/cast v1.3.0
	github.com/zly-app/zapp v1.3.5
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/bridge/opentracing v1.13.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.16.0
	go.opentelemetry.io/otel/sdk v1.16.0
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
