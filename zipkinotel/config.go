package zipkinotel

const (
	defCollectorURL = "http://localhost:9411/api/v2/spans"

	defSamplerFraction = 1

	// 待上传的span队列大小
	defSpanQueueSize = 4096
	// 批次大小
	defSpanBatchSize = 1024
	// 自动旋转时间(秒)
	defAutoRotateTime = 5
	// 上传span超时时间(秒)
	defExportTimeout = 30
)

type Config struct {
	CollectorURL string // 收集器地址, 如 http://localhost:9411/api/v2/spans

	SamplerFraction float64 // 采样器采样率, <= 0.0 表示不采样, 1.0 表示总是采样

	SpanQueueSize        int  // 待上传的span队列大小. 超出的span会被丢弃
	SpanBatchSize        int  // span信息批次发送大小, 存满后一次性发送到jaeger
	BlockOnSpanQueueFull bool // 如果span队列满了, 不会丢弃新的span, 而是阻塞直到有空间. 注意, 开启后如果发生阻塞会影响程序性能.
	AutoRotateTime       int  // 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
	ExportTimeout        int  // 上传span超时时间(秒)
}

func newConfig() *Config {
	return &Config{
		SamplerFraction: defSamplerFraction,
	}
}

func (conf *Config) Check() error {
	if conf.CollectorURL == "" {
		conf.CollectorURL = defCollectorURL
	}

	if conf.SpanBatchSize < 1 {
		conf.SpanBatchSize = defSpanBatchSize
	}
	if conf.SpanQueueSize < 1 {
		conf.SpanQueueSize = defSpanQueueSize
	}
	if conf.SpanQueueSize < conf.SpanBatchSize {
		conf.SpanQueueSize = conf.SpanBatchSize
	}
	if conf.AutoRotateTime < 1 {
		conf.AutoRotateTime = defAutoRotateTime
	}
	if conf.ExportTimeout < 1 {
		conf.ExportTimeout = defExportTimeout
	}
	return nil
}