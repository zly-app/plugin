package prometheus

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/metrics"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/plugin"
)

// 默认组件类型
const PluginType core.PluginType = "metrics"

func init() {
	plugin.RegisterCreatorFunc(PluginType, func(app core.IApp) core.IPlugin {
		conf := newConfig()
		err := app.GetConfig().ParsePluginConfig(PluginType, conf, true)
		if err != nil {
			app.Fatal("解析 metrics 配置失败", zap.Error(err))
		}

		client := NewClient(app, conf)
		metrics.SetClient(client)
		return client
	})
}

// 启用插件
func WithPlugin() zapp.Option {
	return zapp.WithPlugin(PluginType)
}
