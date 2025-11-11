package otlp

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

type logPlugin struct {
	app core.IApp

	setCoreFn func(zapcore.Core)
	conf      *Config

	logProvider *logsdk.LoggerProvider
}

func (p *logPlugin) Init() {
	conf := newConfig()
	err := zapp.App().GetConfig().ParsePluginConfig(DefaultPluginType, conf, true)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		logger.Fatal("解析 otlp 配置失败", zap.Error(err))
	}
	p.app = zapp.App()
	p.conf = conf

	if conf.Log.Enabled {
		p.Log()
	}
}

func (p *logPlugin) Log() {
	compress := otlploghttp.NoCompression
	if p.conf.Trace.Gzip {
		compress = otlploghttp.GzipCompression
	}

	otlpLogOpts := []otlploghttp.Option{
		otlploghttp.WithEndpointURL(p.conf.Log.Addr),
		otlploghttp.WithURLPath(p.conf.Log.URLPath),
		otlploghttp.WithCompression(compress),
		otlploghttp.WithRetry(otlploghttp.RetryConfig{
			Enabled:         p.conf.Log.Retry.Enabled,
			InitialInterval: time.Duration(p.conf.Log.Retry.InitialIntervalSec) * time.Second,
			MaxInterval:     time.Duration(p.conf.Log.Retry.MaxIntervalSec) * time.Second,
			MaxElapsedTime:  time.Duration(p.conf.Log.Retry.MaxElapsedTimeSec) * time.Second,
		}),
	}
	exporter, err := otlploghttp.New(context.Background(), otlpLogOpts...)
	if err != nil {
		p.app.Fatal("无法创建 otlp log 导出器", zap.Error(err))
	}

	batcherOpts := []logsdk.BatchProcessorOption{
		logsdk.WithMaxQueueSize(p.conf.Log.SpanQueueSize),
		logsdk.WithExportMaxBatchSize(p.conf.Log.SpanBatchSize),
		logsdk.WithExportInterval(time.Duration(p.conf.Log.AutoRotateTime) * time.Second),
		logsdk.WithExportTimeout(time.Duration(p.conf.Log.ExportTimeout) * time.Second),
	}

	labels := make([]string, 0, len(p.app.GetConfig().Config().Frame.Labels))
	for k, v := range p.app.GetConfig().Config().Frame.Labels {
		labels = append(labels, k+"="+v)
	}

	lp := logsdk.NewLoggerProvider(
		logsdk.WithProcessor(logsdk.NewBatchProcessor(exporter, batcherOpts...)),
		logsdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(p.app.Name()),
			attribute.Key("app").String(p.app.Name()),
			attribute.String("debug", cast.ToString(p.app.GetConfig().Config().Frame.Debug)),
			attribute.String("env", p.app.GetConfig().Config().Frame.Env),
			attribute.String("flags", strings.Join(p.app.GetConfig().Config().Frame.Flags, ",")),
			attribute.String("labels", strings.Join(labels, ",")),
		)),
	)
	global.SetLoggerProvider(lp)

	p.logProvider = lp

	logCore := otelzap.NewCore(Name, otelzap.WithLoggerProvider(global.GetLoggerProvider()))
	p.setCoreFn(logCore)
}

func (p *logPlugin) Close() {
	if p.logProvider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		_ = p.logProvider.Shutdown(ctx)
	}
}

var defLog *logPlugin
var onceLog sync.Once

func newLogPlugin(setCoreFn func(zapcore.Core)) *logPlugin {
	onceLog.Do(func() {
		defLog = &logPlugin{
			setCoreFn: setCoreFn,
		}
	})
	return defLog
}

func init() {
	zapp.AddHandler(zapp.AfterExitHandler, func(app core.IApp, handlerType zapp.HandlerType) {
		if defLog != nil {
			defLog.Close()
		}
	})
}
