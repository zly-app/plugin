package pprof

const (
	defBind string = ":6060"
)

type Config struct {
	Bind    string // bind http端口
	Disable bool   // 是否关闭pprof
}

func newConfig() *Config {
	return &Config{
		Bind:    defBind,
		Disable: false,
	}
}

func (conf *Config) Check() {
	if conf.Bind == "" {
		conf.Bind = defBind
	}
}
