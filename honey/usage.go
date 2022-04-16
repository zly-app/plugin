package honey

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/plugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 默认插件类型
const DefaultPluginType core.PluginType = "honey"

// 当前服务类型
var nowPluginType = DefaultPluginType

// 设置插件类型, 这个函数应该在 zapp.NewApp 之前调用
func SetPluginType(t core.PluginType) {
	nowPluginType = t
}

// 启用插件
func WithPlugin() zapp.Option {
	plugin.RegisterCreatorFunc(nowPluginType, func(app core.IApp) core.IPlugin {
		return newHoneyPlugin()
	})
	return zapp.WithPlugin(nowPluginType)
}

// 拦截器
func LogHook() zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return newLogInterceptor(core)
	})
}
