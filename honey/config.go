package honey

type Config struct {
	StopLogOutput bool // 停止原有的日志输出
}

func newConfig() *Config {
	return &Config{}
}

func (conf *Config) Check() error {
	return nil
}
