package reporter

import (
	"os"
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
