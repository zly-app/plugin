package pprof

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/plugin"
)

const defPluginType = "pprof"

// 启用插件
func WithPlugin() zapp.Option {
	plugin.RegisterCreatorFunc(defPluginType, func(app core.IApp) core.IPlugin {
		return NewPProf(app)
	})
	return zapp.WithPlugin(defPluginType)
}
