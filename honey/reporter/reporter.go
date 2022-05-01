package reporter

import (
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zly-app/honey/log_data"

	"github.com/zly-app/plugin/honey/component"
)

// 上报者创造者
type IReporterCreator func(c component.IComponent) IReporter

// 上报者
type IReporter interface {
	Start() error
	Close() error
	// 上报
	Report(env, service, instance string, data []*log_data.LogData)
}

var reporterCreators = make(map[string]IReporterCreator)

// 注册上报者创造者
func RegisterReporterCreator(name string, rc IReporterCreator) {
	if _, ok := reporterCreators[name]; ok {
		logger.Log.Fatal("重复注册Reporter创造者", zap.String("name", name))
	}
	reporterCreators[name] = rc
}

// 生成上报者
func MakeReporter(c component.IComponent, name string) IReporter {
	rc, ok := reporterCreators[name]
	if !ok {
		logger.Log.Fatal("试图构建未注册创造者的Reporter", zap.String("name", name))
	}
	return rc(c)
}
