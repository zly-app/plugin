package zipkinotel

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/spf13/cast"
	"github.com/zly-app/zapp/core"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelBridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"go.opentelemetry.io/otel/exporters/zipkin"
)

type ZipkinPlugin struct {
	app      core.IApp
	provider *tracesdk.TracerProvider
}

func NewZipkinPlugin(app core.IApp) core.IPlugin {
	conf := newConfig()
	err := app.GetConfig().ParsePluginConfig(nowPluginType, conf, true)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		app.Fatal("jaeger配置错误", zap.Error(err))
	}

	var exp tracesdk.SpanExporter
	exp, err = zipkin.New(conf.CollectorURL)
	if err != nil {
		app.Fatal("无法创建jaeger跟踪程序", zap.Error(err))
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

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp, batcherOpts...),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(app.Name()),
			attribute.String("debug", cast.ToString(app.GetConfig().Config().Frame.Debug)),
		)),
		tracesdk.WithSampler(
			tracesdk.TraceIDRatioBased(conf.SamplerFraction),
		),
	)
	//otel.SetTracerProvider(tp)

	t := tp.Tracer("")
	bridgeTracer, wrapperTracerProvider := otelBridge.NewTracerPair(t)
	otel.SetTracerProvider(wrapperTracerProvider)
	opentracing.SetGlobalTracer(bridgeTracer)

	return &ZipkinPlugin{
		app:      app,
		provider: tp,
	}
}

func (j *ZipkinPlugin) Inject(a ...interface{}) {}

func (j *ZipkinPlugin) Start() error { return nil }

func (j *ZipkinPlugin) Close() error {
	if j.provider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		return j.provider.Shutdown(ctx)
	}
	return nil
}
