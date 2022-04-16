package reporter

import (
	"os"

	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zly-app/plugin/honey/config"
)

const StdOutReporterName = "stdout"

type StdOutReporter struct{}

func (s *StdOutReporter) Report(logs [][]byte) {
	for _, log := range logs {
		_, _ = os.Stdout.Write(log)
		_, _ = os.Stdout.Write([]byte{'\n'})
	}
}

func NewStdOutReporter() Reporter {
	return &StdOutReporter{}
}

func MakeReporter(conf *config.Config) Reporter {
	switch conf.ReportType {
	case StdOutReporterName:
		return NewStdOutReporter()
	}

	logger.Log.Fatal("honey上报者ReportType未定义", zap.String("ReportType", "conf.ReportType"))
	return nil
}
