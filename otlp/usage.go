package otlp

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/plugin"
	"go.uber.org/zap"
)

// 默认插件类型
const DefaultPluginType core.PluginType = "otlp"

func init() {
	plugin.RegisterCreatorFunc(DefaultPluginType, func(app core.IApp) core.IPlugin {
		conf := newConfig()
		err := app.GetConfig().ParsePluginConfig(DefaultPluginType, conf, true)
		if err != nil {
			app.Fatal("解析 otlp 配置失败", zap.Error(err))
		}

		client := NewOtlpPlugin(app, conf)
		if conf.Metric.Enabled {
			//metrics.SetClient(client)
		}
		return client
	})
}

// 启用插件
func WithPlugin() zapp.Option {
	return zapp.WithPlugin(DefaultPluginType)
}
