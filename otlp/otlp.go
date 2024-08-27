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
	app  core.IApp
	conf *Config

	traceProvider *tracesdk.TracerProvider
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

	ret := &OtlpPlugin{
		app:  app,
		conf: conf,
	}

	ret.Trace()
	return ret
}

func (o *OtlpPlugin) Trace() {
	compress := otlptracehttp.NoCompression
	if o.conf.Trace.Gzip {
		compress = otlptracehttp.GzipCompression
	}

	otlpTraceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(o.conf.Addr),
		otlptracehttp.WithCompression(compress),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         true,
			InitialInterval: 5 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  time.Minute,
		}),
	}
	switch {
	case strings.HasPrefix(o.conf.Addr, "http://"):
		otlpTraceOpts = append(otlpTraceOpts, otlptracehttp.WithInsecure())
		otlpTraceOpts = append(otlpTraceOpts, otlptracehttp.WithEndpoint(strings.TrimPrefix(o.conf.Addr, "http://")))
	case strings.HasPrefix(o.conf.Addr, "https://"):
		otlpTraceOpts = append(otlpTraceOpts, otlptracehttp.WithEndpoint(strings.TrimPrefix(o.conf.Addr, "https://")))
	default:
		otlpTraceOpts = append(otlpTraceOpts, otlptracehttp.WithEndpoint(o.conf.Addr))
	}
	exporter, err := otlptracehttp.New(context.Background(), otlpTraceOpts...)
	if err != nil {
		o.app.Fatal("无法创建otel跟踪程序", zap.Error(err))
	}

	batcherOpts := []tracesdk.BatchSpanProcessorOption{
		tracesdk.WithMaxQueueSize(o.conf.Trace.SpanQueueSize),
		tracesdk.WithMaxExportBatchSize(o.conf.Trace.SpanBatchSize),
		tracesdk.WithBatchTimeout(time.Duration(o.conf.Trace.AutoRotateTime) * time.Second),
		tracesdk.WithExportTimeout(time.Duration(o.conf.Trace.ExportTimeout) * time.Second),
	}
	if o.conf.Trace.BlockOnSpanQueueFull {
		batcherOpts = append(batcherOpts, tracesdk.WithBlocking())
	}

	labels := make([]string, 0, len(o.app.GetConfig().Config().Frame.Labels))
	for k, v := range o.app.GetConfig().Config().Frame.Labels {
		labels = append(labels, k+"="+v)
	}
	sort.Strings(labels)
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter, batcherOpts...),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(o.app.Name()),
			attribute.Key("app").String(o.app.Name()),
			attribute.String("debug", cast.ToString(o.app.GetConfig().Config().Frame.Debug)),
			attribute.String("env", o.app.GetConfig().Config().Frame.Env),
			attribute.String("flags", strings.Join(o.app.GetConfig().Config().Frame.Flags, ",")),
			attribute.String("labels", strings.Join(labels, ",")),
		)),
		tracesdk.WithSampler(
			tracesdk.TraceIDRatioBased(o.conf.Trace.SamplerFraction),
		),
	)

	t := tp.Tracer("")
	bridgeTracer, wrapperTracerProvider := otelBridge.NewTracerPair(t)
	otel.SetTracerProvider(wrapperTracerProvider)
	opentracing.SetGlobalTracer(bridgeTracer)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	o.traceProvider = tp
}

func (o *OtlpPlugin) Inject(a ...interface{}) {}

func (o *OtlpPlugin) Start() error { return nil }

func (o *OtlpPlugin) Close() error {
	if o.traceProvider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		return o.traceProvider.Shutdown(ctx)
	}
	return nil
}
