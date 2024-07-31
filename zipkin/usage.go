package zipkin

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/plugin"
)

// 默认插件类型
const DefaultPluginType core.PluginType = "zipkin"

func init() {
	plugin.RegisterCreatorFunc(DefaultPluginType, func(app core.IApp) core.IPlugin {
		return NewZipKinPlugin(app)
	})
}

// 启用插件
func WithPlugin() zapp.Option {
	return zapp.WithPlugin(DefaultPluginType)
}
