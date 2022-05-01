package std

import (
	"os"
	"strings"
	"time"

	"github.com/zly-app/honey/log_data"

	"github.com/zly-app/plugin/honey/component"
	"github.com/zly-app/plugin/honey/reporter"
)

var StdFormat = "[{dev}.{service}][{instance}][{time}] {level} {msg} {fields} {line} {trace_id}"
var TimeFormat = "2006-01-02 15:04:05.999999"

type StdReporter struct{}

func (s *StdReporter) Start() error { return nil }
func (s *StdReporter) Close() error { return nil }

func (s *StdReporter) Report(env, service, instance string, data []*log_data.LogData) {
	for _, v := range data {
		text := StdFormat
		text = strings.ReplaceAll(text, "{dev}", env)
		text = strings.ReplaceAll(text, "{service}", service)
		text = strings.ReplaceAll(text, "{instance}", instance)
		text = strings.ReplaceAll(text, "{time}", time.Unix(0, v.T*1000).Format(TimeFormat))
		text = strings.ReplaceAll(text, "{level}", v.Level)
		text = strings.ReplaceAll(text, "{msg}", v.Msg)
		text = strings.ReplaceAll(text, "{fields}", v.Fields)
		text = strings.ReplaceAll(text, "{line}", v.Line)
		text = strings.ReplaceAll(text, "{trace_id}", v.TraceID)

		_, _ = os.Stdout.WriteString(text)
		_, _ = os.Stdout.Write([]byte{'\n'})
	}
}

const StdReporterName = "std"

func init() {
	reporter.RegisterReporterCreator(StdReporterName, func(c component.IComponent) reporter.IReporter {
		return &StdReporter{}
	})
}
