package otlp

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log/global"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
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
		log.Fatal("解析 otlp 配置失败", zap.Error(err))
	}
	p.app = zapp.App()
	p.conf = conf

	if conf.Log.Enabled {
		p.Log()
	}
}

func (p *logPlugin) Log() {
	exporter, err := newLogExporter(p.conf.Log, p.conf.effectiveProtocol(p.conf.Log.Protocol), p.conf.mergeHeaders(p.conf.Log.Headers))
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

	pid := os.Getpid()
	lp := logsdk.NewLoggerProvider(
		logsdk.WithProcessor(logsdk.NewBatchProcessor(exporter, batcherOpts...)),
		logsdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(p.app.Name()),
			attribute.Key("app").String(p.app.Name()),
			attribute.String("debug", cast.ToString(p.app.GetConfig().Config().Frame.Debug)),
			attribute.String("env", p.app.GetConfig().Config().Frame.Env),
			attribute.String("instance", cast.ToString(p.app.GetConfig().Config().Frame.Instance)),
			attribute.String("flags", strings.Join(p.app.GetConfig().Config().Frame.Flags, ",")),
			attribute.String("labels", strings.Join(labels, ",")),
			attribute.Int("process.pid", pid),
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
