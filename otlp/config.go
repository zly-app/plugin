package otlp

const (
	defAddr = "http://localhost:4318"

	defTraceSamplerFraction      = 1
	defTraceSpanQueueSize        = 8192
	defTraceSpanBatchSize        = 2048
	defTraceBlockOnSpanQueueFull = false
	defTraceAutoRotateTime       = 5
	defTraceExportTimeout        = 30
	defTraceGzip                 = false
)

type TraceConfig struct {
	SamplerFraction      float64 // 采样器采样率, <= 0.0 表示不采样, 1.0 表示总是采样
	SpanQueueSize        int     // 待上传的span队列大小. 超出的span会被丢弃
	SpanBatchSize        int     // span信息批次发送大小, 存满后一次性发送到jaeger
	BlockOnSpanQueueFull bool    // 如果span队列满了, 不会丢弃新的span, 而是阻塞直到有空间. 注意, 开启后如果发生阻塞会影响程序性能.
	AutoRotateTime       int     // 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
	ExportTimeout        int     // 上传span超时时间(秒)
	Gzip                 bool    // 是否启用gzip压缩
}
type Config struct {
	Addr         string // 地址, 如 http://localhost:4318
	ProxyAddress string // 代理地址. 支持 http, https, socks5, socks5h. 示例: socks5://127.0.0.1:1080 socks5://user:pwd@127.0.0.1:1080

	Trace TraceConfig // trace 配置
}

func newConfig() *Config {
	return &Config{
		Trace: TraceConfig{
			SamplerFraction:      defTraceSamplerFraction,
			SpanQueueSize:        defTraceSpanQueueSize,
			SpanBatchSize:        defTraceSpanBatchSize,
			BlockOnSpanQueueFull: defTraceBlockOnSpanQueueFull,
			AutoRotateTime:       defTraceAutoRotateTime,
			ExportTimeout:        defTraceExportTimeout,
			Gzip:                 defTraceGzip,
		},
	}
}

func (conf *Config) Check() error {
	if conf.Addr == "" {
		conf.Addr = defAddr
	}

	if conf.Trace.SpanBatchSize < 1 {
		conf.Trace.SpanBatchSize = defTraceSpanBatchSize
	}
	if conf.Trace.SpanQueueSize < 1 {
		conf.Trace.SpanQueueSize = defTraceSpanQueueSize
	}
	if conf.Trace.SpanQueueSize < conf.Trace.SpanBatchSize {
		conf.Trace.SpanQueueSize = conf.Trace.SpanBatchSize
	}
	if conf.Trace.AutoRotateTime < 1 {
		conf.Trace.AutoRotateTime = defTraceAutoRotateTime
	}
	if conf.Trace.ExportTimeout < 1 {
		conf.Trace.ExportTimeout = defTraceExportTimeout
	}
	return nil
}
