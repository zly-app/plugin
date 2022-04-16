package reporter

import (
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zly-app/plugin/honey/config"
)

// 上报者
type Reporter interface {
	// 上报
	Report(logs [][]byte)
}

func MakeReporter(conf *config.Config) Reporter {
	switch conf.ReportType {
	case StdOutReporterName:
		return NewStdOutReporter()
	}

	logger.Log.Fatal("honey上报者类型未定义", zap.String("ReportType", conf.ReportType))
	return nil
}
