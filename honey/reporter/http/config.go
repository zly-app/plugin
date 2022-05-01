package http

const (
	// 默认上报地址
	defReportAddress = "http://127.0.0.1:8080/report"
	// 默认压缩器名
	defCompress = "zstd"
	// 默认序列化器名
	defaultSerializer = "msgpack"
	// 默认上报超时
	defReportTimeout = 5
)

type Config struct {
	Disable       bool   // 关闭
	ReportAddress string // 上报地址, 示例: http://127.0.0.1:8080/report
	Compress      string // 压缩器名
	Serializer    string // 序列化器名
	AuthToken     string // 验证token, 如何设置, 请求header必须带上 token={AuthToken}, 如 token=myAuthToken
	ReportTimeout int    // 上报超时, 单位秒
}

func newConfig() *Config {
	return &Config{}
}

func (c *Config) Check() error {
	if c.ReportAddress == "" {
		c.ReportAddress = defReportAddress
	}
	if c.Compress == "" {
		c.Compress = defCompress
	}
	if c.Serializer == "" {
		c.Serializer = defaultSerializer
	}
	if c.ReportTimeout < 1 {
		c.ReportTimeout = defReportTimeout
	}
	return nil
}
