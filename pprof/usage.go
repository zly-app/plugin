package pprof

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/plugin"
)

const defPluginType = "pprof"

func init() {
	plugin.RegisterCreatorFunc(defPluginType, func(app core.IApp) core.IPlugin {
		return NewPProf(app)
	})
}

// 启用插件
func WithPlugin() zapp.Option {
	return zapp.WithPlugin(defPluginType)
}
