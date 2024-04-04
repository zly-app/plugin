package pprof

const (
	defBind   = ":6060"
	defEnable = false
)

type Config struct {
	Bind   string // bind http端口
	Enable bool   // 是否开启pprof
}

func newConfig() *Config {
	return &Config{
		Bind:   defBind,
		Enable: defEnable,
	}
}

func (conf *Config) Check() {
	if conf.Bind == "" {
		conf.Bind = defBind
	}
}
