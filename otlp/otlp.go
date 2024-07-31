package otlp

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/spf13/cast"
	"github.com/zly-app/zapp/core"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelBridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
)

type OtlpPlugin struct {
	app      core.IApp
	provider *tracesdk.TracerProvider
}

func NewOtlpPlugin(app core.IApp) core.IPlugin {
	conf := newConfig()
	err := app.GetConfig().ParsePluginConfig(DefaultPluginType, conf, true)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		app.Fatal("otlp配置错误", zap.Error(err))
	}

	otlpTraceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(conf.Addr),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         true,
			InitialInterval: 5 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  time.Minute,
		}),
	}
	switch {
	case strings.HasPrefix(conf.Addr, "http://"):
		otlpTraceOpts = append(otlpTraceOpts, otlptracehttp.WithInsecure())
		otlpTraceOpts = append(otlpTraceOpts, otlptracehttp.WithEndpoint(strings.TrimPrefix(conf.Addr, "http://")))
	case strings.HasPrefix(conf.Addr, "https://"):
		otlpTraceOpts = append(otlpTraceOpts, otlptracehttp.WithEndpoint(strings.TrimPrefix(conf.Addr, "https://")))
	default:
		otlpTraceOpts = append(otlpTraceOpts, otlptracehttp.WithEndpoint(conf.Addr))
	}
	exporter, err := otlptracehttp.New(context.Background(), otlpTraceOpts...)
	if err != nil {
		app.Fatal("无法创建otel跟踪程序", zap.Error(err))
	}

	batcherOpts := []tracesdk.BatchSpanProcessorOption{
		tracesdk.WithMaxQueueSize(conf.SpanQueueSize),
		tracesdk.WithMaxExportBatchSize(conf.SpanBatchSize),
		tracesdk.WithBatchTimeout(time.Duration(conf.AutoRotateTime) * time.Second),
		tracesdk.WithExportTimeout(time.Duration(conf.ExportTimeout) * time.Second),
	}
	if conf.BlockOnSpanQueueFull {
		batcherOpts = append(batcherOpts, tracesdk.WithBlocking())
	}

	labels := make([]string, 0, len(app.GetConfig().Config().Frame.Labels))
	for k, v := range app.GetConfig().Config().Frame.Labels {
		labels = append(labels, k+"="+v)
	}
	sort.Strings(labels)
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter, batcherOpts...),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(app.Name()),
			attribute.Key("app").String(app.Name()),
			attribute.String("debug", cast.ToString(app.GetConfig().Config().Frame.Debug)),
			attribute.String("env", app.GetConfig().Config().Frame.Env),
			attribute.StringSlice("flags", app.GetConfig().Config().Frame.Flags),
			attribute.StringSlice("labels", labels),
		)),
		tracesdk.WithSampler(
			tracesdk.TraceIDRatioBased(conf.SamplerFraction),
		),
	)

	t := tp.Tracer("")
	bridgeTracer, wrapperTracerProvider := otelBridge.NewTracerPair(t)
	otel.SetTracerProvider(wrapperTracerProvider)
	opentracing.SetGlobalTracer(bridgeTracer)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &OtlpPlugin{
		app:      app,
		provider: tp,
	}
}

func (j *OtlpPlugin) Inject(a ...interface{}) {}

func (j *OtlpPlugin) Start() error { return nil }

func (j *OtlpPlugin) Close() error {
	if j.provider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		return j.provider.Shutdown(ctx)
	}
	return nil
}
