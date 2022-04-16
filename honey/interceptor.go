package honey

import (
	"go.uber.org/zap/zapcore"
)

type interceptor struct {
	zapcore.Core
	honey *HoneyPlugin
}

func newLogInterceptor(core zapcore.Core) zapcore.Core {
	return &interceptor{
		Core:  core,
		honey: newHoneyPlugin(),
	}
}

func (c *interceptor) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}
func (h *interceptor) With(fields []zapcore.Field) zapcore.Core {
	return &interceptor{
		Core: h.Core.With(fields),
	}
}
func (c *interceptor) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	if c.honey.Interceptor(&ent, fields) {
		return nil
	}
	return c.Core.Write(ent, fields)
}
