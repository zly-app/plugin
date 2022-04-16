package honey

import (
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type HoneyPlugin struct {
	conf *Config
}

func newHoneyPlugin() *HoneyPlugin {
	conf := newConfig()

	// 解析配置
	err := config.Conf.ParsePluginConfig(nowPluginType, conf)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		logger.Log.Panic("honey配置错误", zap.String("nowPluginType", string(nowPluginType)), zap.Error(err))
	}
	return &HoneyPlugin{
		conf: conf,
	}
}

func (h *HoneyPlugin) Interceptor(ent *zapcore.Entry, fields []zapcore.Field) (cancel bool) {
	return h.conf.StopLogOutput
}
