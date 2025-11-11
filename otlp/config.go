package otlp

const (
	defRetryEnabled            = true
	defRetryInitialIntervalSec = 5
	defRetryMaxIntervalSec     = 30
	defRetryMaxElapsedTimeSec  = 60

	defTraceEnabled              = true
	defTraceAddr                 = "http://localhost:4318"
	defTraceURLPath              = "/v1/traces"
	defTraceGzip                 = true
	defTraceSamplerFraction      = 1
	defTraceSpanQueueSize        = 8192
	defTraceSpanBatchSize        = 2048
	defTraceBlockOnSpanQueueFull = false
	defTraceAutoRotateTime       = 5
	defTraceExportTimeout        = 30

	defMetricEnabled        = true
	defMetricAddr           = "http://localhost:4318"
	defMetricURLPath        = "/v1/metrics"
	defMetricGzip           = true
	defMetricAutoRotateTime = 5
	defMetricExportTimeout  = 30
)

type RetryConfig struct {
	Enabled            bool  // 是否启用
	InitialIntervalSec int64 // 第一次上传失败的重试间隔秒数
	MaxIntervalSec     int64 // 最大重试间隔秒数
	MaxElapsedTimeSec  int64 // 超过这个秒数后则放弃这一批数据
}
type TraceConfig struct {
	Enabled bool   // 是否启用
	Addr    string // 地址, 如 http://localhost:4318
	URLPath string // path, 如 /v1/traces
	Gzip    bool   // 是否启用gzip压缩

	SamplerFraction      float64 // 采样器采样率, <= 0.0 表示不采样, 1.0 表示总是采样
	SpanQueueSize        int     // 待上传的span队列大小. 超出的span会被丢弃
	SpanBatchSize        int     // span信息批次发送大小, 存满后一次性发送到jaeger
	BlockOnSpanQueueFull bool    // 如果span队列满了, 不会丢弃新的span, 而是阻塞直到有空间. 注意, 开启后如果发生阻塞会影响程序性能.
	AutoRotateTime       int     // 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
	ExportTimeout        int     // 上传span超时时间(秒)

	Retry RetryConfig // 重试配置
}
type MetricConfig struct {
	Enabled bool   // 是否启用
	Addr    string // 地址, 如 http://localhost:4318
	URLPath string // path, 如 /v1/metrics
	Gzip    bool   // 是否启用gzip压缩

	AutoRotateTime int // 自动旋转时间(秒)
	ExportTimeout  int // 上传metric超时时间(秒)

	Retry RetryConfig // 重试配置
}

type Config struct {
	Trace  TraceConfig  // trace 配置
	Metric MetricConfig // metric 配置
}

func newConfig() *Config {
	return &Config{
		Trace: TraceConfig{
			Enabled:              defTraceEnabled,
			Addr:                 defTraceAddr,
			URLPath:              defMetricURLPath,
			Gzip:                 defTraceGzip,
			SamplerFraction:      defTraceSamplerFraction,
			SpanQueueSize:        defTraceSpanQueueSize,
			SpanBatchSize:        defTraceSpanBatchSize,
			BlockOnSpanQueueFull: defTraceBlockOnSpanQueueFull,
			AutoRotateTime:       defTraceAutoRotateTime,
			ExportTimeout:        defTraceExportTimeout,
			Retry: RetryConfig{
				Enabled:            defRetryEnabled,
				InitialIntervalSec: defRetryInitialIntervalSec,
				MaxIntervalSec:     defRetryMaxIntervalSec,
				MaxElapsedTimeSec:  defRetryMaxElapsedTimeSec,
			},
		},
		Metric: MetricConfig{
			Enabled:        defMetricEnabled,
			Addr:           defMetricAddr,
			URLPath:        defMetricURLPath,
			Gzip:           defMetricGzip,
			AutoRotateTime: defMetricAutoRotateTime,
			ExportTimeout:  defMetricExportTimeout,
			Retry: RetryConfig{
				Enabled:            defRetryEnabled,
				InitialIntervalSec: defRetryInitialIntervalSec,
				MaxIntervalSec:     defRetryMaxIntervalSec,
				MaxElapsedTimeSec:  defRetryMaxElapsedTimeSec,
			},
		},
	}
}

func (conf *Config) Check() error {
	if conf.Trace.Addr == "" {
		conf.Trace.Addr = defTraceAddr
	}
	if conf.Trace.URLPath == "" {
		conf.Trace.URLPath = defTraceURLPath
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
	if conf.Trace.Retry.InitialIntervalSec < 1 {
		conf.Trace.Retry.InitialIntervalSec = defRetryInitialIntervalSec
	}
	if conf.Trace.Retry.MaxIntervalSec < 1 {
		conf.Trace.Retry.MaxIntervalSec = defRetryMaxIntervalSec
	}
	if conf.Trace.Retry.MaxElapsedTimeSec < 1 {
		conf.Trace.Retry.MaxElapsedTimeSec = defRetryMaxElapsedTimeSec
	}

	if conf.Metric.Addr == "" {
		conf.Metric.Addr = defMetricAddr
	}
	if conf.Metric.URLPath == "" {
		conf.Metric.URLPath = defMetricURLPath
	}
	if conf.Metric.AutoRotateTime < 1 {
		conf.Metric.AutoRotateTime = defMetricAutoRotateTime
	}
	if conf.Metric.ExportTimeout < 1 {
		conf.Metric.ExportTimeout = defMetricAutoRotateTime
	}
	if conf.Metric.Retry.InitialIntervalSec < 1 {
		conf.Metric.Retry.InitialIntervalSec = defRetryInitialIntervalSec
	}
	if conf.Metric.Retry.MaxIntervalSec < 1 {
		conf.Metric.Retry.MaxIntervalSec = defRetryMaxIntervalSec
	}
	if conf.Metric.Retry.MaxElapsedTimeSec < 1 {
		conf.Metric.Retry.MaxElapsedTimeSec = defRetryMaxElapsedTimeSec
	}

	return nil
}
