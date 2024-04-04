package pprof

const (
	defBind    = ":6060"
	defDisable = true
)

type Config struct {
	Bind    string // bind http端口
	Disable bool   // 是否关闭pprof
}

func newConfig() *Config {
	return &Config{
		Bind:    defBind,
		Disable: defDisable,
	}
}

func (conf *Config) Check() {
	if conf.Bind == "" {
		conf.Bind = defBind
	}
}
