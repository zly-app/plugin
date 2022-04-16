package config

const (
	// 默认最大缓存长度
	DefaultCacheLen = 10000
	// 默认批次长度
	DefaultBatchLen = 1000
	// 默认等待上报间隔时间(秒)
	DefaultWaitReport = 3
	// 默认上报重试次数
	DefaultReportRetryCount = 2
	// 默认压缩类型
	DefaultCompressType = "zstd"

	// 默认上报者类型
	DefaultReportType = "stdout"
)

type Config struct {
	Service  string // 服务名
	Instance string // 实例名

	CacheLen         int  // 最大缓存长度, 超出这个长度时日志会等待处理完毕才会继续往下执行
	BatchLen         int  // 批次长度, 累计达到这个长度立即上报一次日志, 不用等待时间
	WaitReport       int  // 默认等待上报间隔时间(秒), 如果没有达到累计上报长度, 在指定时间后也会上报
	ReportRetryCount int  // 上报重试次数
	StopLogOutput    bool // 停止原有的日志输出

	CompressType string // 压缩类型, 支持 zstd

	ReportType string // 上报者类型, 支持 stdout, http
	HttpReport struct {
		Address string // 上报地址

	} // http上报者
}

func NewConfig() *Config {
	return &Config{
		ReportRetryCount: DefaultReportRetryCount,
	}
}

func (conf *Config) Check() error {
	if conf.CacheLen <= 0 {
		conf.CacheLen = DefaultCacheLen
	}
	if conf.BatchLen <= 0 {
		conf.BatchLen = DefaultBatchLen
	}
	if conf.WaitReport <= 0 {
		conf.WaitReport = DefaultWaitReport
	}
	if conf.Instance == "" {

	}
	if conf.CompressType == "" {
		conf.CompressType = DefaultCompressType
	}
	if conf.ReportType == "" {
		conf.ReportType = DefaultReportType
	}
	return nil
}
