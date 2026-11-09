package honey

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/zlog"

	_ "github.com/zly-app/honey/output/honey-http"
	_ "github.com/zly-app/honey/output/loki-http"
	_ "github.com/zly-app/honey/output/std"
)

// DefaultPluginType 默认插件类型
const DefaultPluginType core.PluginType = "honey"

// WithPlugin 启用插件，必须在 zapp 初始化前调用。
func WithPlugin() zapp.Option {
	h := NewHoneyPlugin()
	logConf := zlog.NewHookConfig().
		AddStartHookCallbacks(h.Init).           // 通过日志启动hook的能力提供初始化
		AddInterceptorFunc(h.LogInterceptorFunc) // 添加拦截函数
	return zapp.WithMultiOptions(
		zapp.WithLoggerOptions(zlog.WithHookByConfig(logConf)),
		zapp.WithHandler(zapp.AfterInitializeHandler, func(app core.IApp, handlerType zapp.HandlerType) {
			NewHoneyPlugin().Start()
		}),
		zapp.WithHandler(zapp.AfterStartHandler, func(app core.IApp, handlerType zapp.HandlerType) {
			NewHoneyPlugin().OnAppStart()
		}),
		zapp.WithHandler(zapp.BeforeExitHandler, func(app core.IApp, handlerType zapp.HandlerType) {
			NewHoneyPlugin().BeforeClose()
		}),
		zapp.WithHandler(zapp.AfterExitHandler, func(app core.IApp, handlerType zapp.HandlerType) {
			NewHoneyPlugin().AfterClose()
		}),
	)
}
