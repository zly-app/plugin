package jaeger

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/plugin"
)

// DefaultPluginType 默认插件类型
const DefaultPluginType core.PluginType = "jaeger"

// WithPlugin 启用插件，必须在 zapp 初始化前调用。
func WithPlugin() zapp.Option {
	plugin.RegisterCreatorFunc(DefaultPluginType, func(app core.IApp) core.IPlugin {
		return NewJaegerPlugin(app)
	})
	return zapp.WithPlugin(DefaultPluginType)
}
