package otlp

import (
	zapp_metrics "github.com/zly-app/zapp/component/metrics"
	"go.uber.org/zap"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/zlog"
	"github.com/zly-app/zapp/plugin"

	"github.com/zly-app/plugin/otlp/metrics"
)

// DefaultPluginType 默认插件类型
const DefaultPluginType core.PluginType = "otlp"

// WithPlugin 启用插件，必须在 zapp 初始化前调用。
func WithPlugin() zapp.Option {
	plugin.RegisterCreatorFunc(DefaultPluginType, func(app core.IApp) core.IPlugin {
		conf := newConfig()
		err := app.GetConfig().ParsePluginConfig(DefaultPluginType, conf, true)
		if err != nil {
			app.Fatal("解析 otlp 配置失败", zap.Error(err))
		}

		client := NewOtlpPlugin(app, conf)
		if conf.Metric.Enabled {
			zapp_metrics.SetClient(metrics.NewClient(client.metricMeter))
		}
		return client
	})

	logConf := zlog.NewHookConfig()
	h := newLogPlugin(logConf.SetCore)
	logConf.AddStartHookCallbacks(h.Init) // 通过日志启动hook的能力提供初始化

	return zapp.WithMultiOptions(
		zapp.WithPlugin(DefaultPluginType),
		zapp.WithLoggerOptions(zlog.WithHookByConfig(logConf)),
		zapp.WithHandler(zapp.AfterExitHandler, func(app core.IApp, handlerType zapp.HandlerType) {
			if defLog != nil {
				defLog.Close()
			}
		}),
	)
}
