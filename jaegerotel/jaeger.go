package jaegerotel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/spf13/cast"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelBridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"go.opentelemetry.io/otel/exporters/jaeger"
)

type JaegerPlugin struct {
	app      core.IApp
	provider *tracesdk.TracerProvider
}

func NewJaegerPlugin(app core.IApp) core.IPlugin {
	conf := newConfig()
	err := app.GetConfig().ParsePluginConfig(nowPluginType, conf, true)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		app.Fatal("jaeger配置错误", zap.Error(err))
	}

	transport := &http.Transport{
		MaxIdleConns:        5,                // 最大连接数
		MaxIdleConnsPerHost: 2,                // 最大空闲连接数
		IdleConnTimeout:     time.Second * 30, // 空闲连接在关闭自己之前保持空闲的最大时间
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过tls校验
			RootCAs:            x509.NewCertPool(),
		},
	}
	if conf.ProxyAddress != "" {
		p, err := utils.NewHttpProxy(conf.ProxyAddress)
		if err != nil {
			logger.Log.Fatal("创建loki-http代理失败", zap.Error(err))
		}
		p.SetProxy(transport)
	}
	client := &http.Client{
		Transport: transport,
	}
	var exp tracesdk.SpanExporter
	if conf.Endpoint != "" {
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(conf.Endpoint),
			jaeger.WithUsername(conf.User),
			jaeger.WithPassword(conf.Password),
			jaeger.WithHTTPClient(client),
		))
	} else {
		exp, err = jaeger.New(jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(conf.AgentHost),
			jaeger.WithAgentPort(cast.ToString(conf.AgentPort)),
		))
	}
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
			attribute.Key("app").String(app.Name()),
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

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &JaegerPlugin{
		app:      app,
		provider: tp,
	}
}

func (j *JaegerPlugin) Inject(a ...interface{}) {}

func (j *JaegerPlugin) Start() error { return nil }

func (j *JaegerPlugin) Close() error {
	if j.provider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		return j.provider.Shutdown(ctx)
	}
	return nil
}
