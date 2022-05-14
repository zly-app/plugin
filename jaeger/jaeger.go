package jaeger

import (
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/zly-app/zapp/core"
	"go.uber.org/zap"

	"github.com/uber/jaeger-client-go/config"
)

type JaegerPlugin struct {
	app    core.IApp
	closer io.Closer
}

func NewJaegerPlugin(app core.IApp) core.IPlugin {
	conf := newConfig()

	// 解析配置
	key := "plugins." + string(nowPluginType)
	vi := app.GetConfig().GetViper()
	if vi.IsSet(key) {
		err := vi.UnmarshalKey(key, conf)
		if err != nil {
			app.Fatal("无法解析插件配置", zap.String("PluginType", string(nowPluginType)), zap.Error(err))
		}
	}

	if err := conf.Check(); err != nil {
		app.Fatal("jaeger配置错误", zap.Error(err))
	}

	cfg := &config.Configuration{
		ServiceName: app.Name(),
		Sampler: &config.SamplerConfig{
			Type:  conf.SamplerType,
			Param: conf.SamplerParam,
		},
		Reporter: &config.ReporterConfig{
			QueueSize:           conf.SpanBatchSize,
			BufferFlushInterval: time.Duration(conf.AutoRotateTime) * time.Second,
			LogSpans:            true, // 推送log信息
			LocalAgentHostPort:  conf.Address,
			User:                conf.User,
			Password:            conf.Password,
		},
	}
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		app.Fatal("无法创建jaeger跟踪程序", zap.Error(err))
	}
	opentracing.SetGlobalTracer(tracer)

	return &JaegerPlugin{
		app:    app,
		closer: closer,
	}
}

func (j *JaegerPlugin) Inject(a ...interface{}) {}

func (j *JaegerPlugin) Start() error { return nil }

func (j *JaegerPlugin) Close() error {
	if j.closer != nil {
		return j.closer.Close()
	}
	return nil
}
