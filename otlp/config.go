package otlp

import (
	"fmt"
	"strings"
)

// OTLP 传输协议, 取值与 OpenTelemetry 规范的 OTEL_EXPORTER_OTLP_PROTOCOL 一致
const (
	// ProtocolHTTPProtobuf OTLP/HTTP, protobuf 载荷, 默认端口 4318
	ProtocolHTTPProtobuf = "http/protobuf"
	// ProtocolGRPC OTLP/gRPC, 默认端口 4317
	ProtocolGRPC = "grpc"
)

const (
	defRetryEnabled            = true
	defRetryInitialIntervalSec = 5
	defRetryMaxIntervalSec     = 30
	defRetryMaxElapsedTimeSec  = 60

	defProtocol = ProtocolGRPC
	defHTTPAddr = "http://localhost:4318"
	defGRPCAddr = "http://localhost:4317"

	defLogEnabled        = true
	defLogURLPath        = "/v1/logs"
	defLogGzip           = true
	defLogSpanQueueSize  = 8192
	defLogSpanBatchSize  = 2048
	defLogAutoRotateTime = 5
	defLogExportTimeout  = 30

	defTraceEnabled              = true
	defTraceURLPath              = "/v1/traces"
	defTraceGzip                 = true
	defTraceSamplerFraction      = 1
	defTraceSpanQueueSize        = 8192
	defTraceSpanBatchSize        = 2048
	defTraceBlockOnSpanQueueFull = false
	defTraceAutoRotateTime       = 5
	defTraceExportTimeout        = 30

	defMetricEnabled        = true
	defMetricURLPath        = "/v1/metrics"
	defMetricGzip           = true
	defProcessCollector     = true
	defGoCollector          = true
	defMetricAutoRotateTime = 5
	defMetricExportTimeout  = 30
)

type RetryConfig struct {
	Enabled            bool  // 是否启用
	InitialIntervalSec int64 // 第一次上传失败的重试间隔秒数
	MaxIntervalSec     int64 // 最大重试间隔秒数
	MaxElapsedTimeSec  int64 // 超过这个秒数后则放弃这一批数据
}
type LogConfig struct {
	Enabled  bool              // 是否启用
	Protocol string            // 传输协议, 可选 grpc / http/protobuf, 为空时使用全局 Protocol
	Addr     string            // 地址, 如 http://localhost:4318(HTTP) 或 localhost:4317(gRPC)
	URLPath  string            // path, 如 /v1/logs. 仅 http/protobuf 生效
	Gzip     bool              // 是否启用gzip压缩
	Headers  map[string]string // 附加请求头(如认证信息), 与全局 Headers 合并, 同名覆盖

	SpanQueueSize  int // 待上传的span队列大小. 超出的log会被丢弃
	SpanBatchSize  int // log信息批次发送大小, 存满后一次性发送
	AutoRotateTime int // 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
	ExportTimeout  int // 上传span超时时间(秒)

	Retry RetryConfig // 重试配置
}
type TraceConfig struct {
	Enabled  bool              // 是否启用
	Protocol string            // 传输协议, 可选 grpc / http/protobuf, 为空时使用全局 Protocol
	Addr     string            // 地址, 如 http://localhost:4318(HTTP) 或 localhost:4317(gRPC)
	URLPath  string            // path, 如 /v1/traces. 仅 http/protobuf 生效
	Gzip     bool              // 是否启用gzip压缩
	Headers  map[string]string // 附加请求头(如认证信息), 与全局 Headers 合并, 同名覆盖

	SamplerFraction      float64 // 采样器采样率, <= 0.0 表示不采样, 1.0 表示总是采样
	SpanQueueSize        int     // 待上传的span队列大小. 超出的span会被丢弃
	SpanBatchSize        int     // span信息批次发送大小, 存满后一次性发送
	BlockOnSpanQueueFull bool    // 如果span队列满了, 不会丢弃新的span, 而是阻塞直到有空间. 注意, 开启后如果发生阻塞会影响程序性能.
	AutoRotateTime       int     // 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
	ExportTimeout        int     // 上传span超时时间(秒)

	Retry RetryConfig // 重试配置
}
type MetricConfig struct {
	Enabled  bool              // 是否启用
	Protocol string            // 传输协议, 可选 grpc / http/protobuf, 为空时使用全局 Protocol
	Addr     string            // 地址, 如 http://localhost:4318(HTTP) 或 localhost:4317(gRPC)
	URLPath  string            // path, 如 /v1/metrics. 仅 http/protobuf 生效
	Gzip     bool              // 是否启用gzip压缩
	Headers  map[string]string // 附加请求头(如认证信息), 与全局 Headers 合并, 同名覆盖

	ProcessCollector bool // 启用进程收集器
	GoCollector      bool // 启用go收集器

	AutoRotateTime int // 自动旋转时间(秒)
	ExportTimeout  int // 上传metric超时时间(秒)

	Retry RetryConfig // 重试配置
}

type Config struct {
	Protocol string            // 全局默认传输协议, 可选 grpc / http/protobuf, 为空时使用 grpc
	Headers  map[string]string // 全局附加请求头(如认证信息), 各信号可覆盖
	Log      LogConfig         // log 配置
	Trace    TraceConfig       // trace 配置
	Metric   MetricConfig
}

func newConfig() *Config {
	return &Config{
		Protocol: defProtocol,
		Log: LogConfig{
			Enabled:        defLogEnabled,
			URLPath:        defLogURLPath,
			Gzip:           defLogGzip,
			SpanQueueSize:  defLogSpanQueueSize,
			SpanBatchSize:  defLogSpanBatchSize,
			AutoRotateTime: defLogAutoRotateTime,
			ExportTimeout:  defLogExportTimeout,
			Retry: RetryConfig{
				Enabled:            defRetryEnabled,
				InitialIntervalSec: defRetryInitialIntervalSec,
				MaxIntervalSec:     defRetryMaxIntervalSec,
				MaxElapsedTimeSec:  defRetryMaxElapsedTimeSec,
			},
		},
		Trace: TraceConfig{
			Enabled:              defTraceEnabled,
			URLPath:              defTraceURLPath,
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
			Enabled:          defMetricEnabled,
			URLPath:          defMetricURLPath,
			Gzip:             defMetricGzip,
			ProcessCollector: defProcessCollector,
			GoCollector:      defGoCollector,
			AutoRotateTime:   defMetricAutoRotateTime,
			ExportTimeout:    defMetricExportTimeout,
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
	var err error
	if conf.Protocol, err = normalizeProtocol(conf.Protocol); err != nil {
		return err
	}
	if conf.Log.Protocol, err = normalizeProtocol(conf.Log.Protocol); err != nil {
		return fmt.Errorf("Log 配置错误: %w", err)
	}
	if conf.Trace.Protocol, err = normalizeProtocol(conf.Trace.Protocol); err != nil {
		return fmt.Errorf("Trace 配置错误: %w", err)
	}
	if conf.Metric.Protocol, err = normalizeProtocol(conf.Metric.Protocol); err != nil {
		return fmt.Errorf("Metric 配置错误: %w", err)
	}

	if conf.Log.Addr == "" {
		conf.Log.Addr = defAddr(conf.effectiveProtocol(conf.Log.Protocol))
	}
	if conf.Log.URLPath == "" {
		conf.Log.URLPath = defLogURLPath
	}
	if conf.Log.SpanBatchSize < 1 {
		conf.Log.SpanBatchSize = defLogSpanBatchSize
	}
	if conf.Log.SpanQueueSize < 1 {
		conf.Log.SpanQueueSize = defLogSpanQueueSize
	}
	if conf.Log.SpanQueueSize < conf.Log.SpanBatchSize {
		conf.Log.SpanQueueSize = conf.Log.SpanBatchSize
	}
	if conf.Log.AutoRotateTime < 1 {
		conf.Log.AutoRotateTime = defLogAutoRotateTime
	}
	if conf.Log.ExportTimeout < 1 {
		conf.Log.ExportTimeout = defLogExportTimeout
	}
	if conf.Log.Retry.InitialIntervalSec < 1 {
		conf.Log.Retry.InitialIntervalSec = defRetryInitialIntervalSec
	}
	if conf.Log.Retry.MaxIntervalSec < 1 {
		conf.Log.Retry.MaxIntervalSec = defRetryMaxIntervalSec
	}
	if conf.Log.Retry.MaxElapsedTimeSec < 1 {
		conf.Log.Retry.MaxElapsedTimeSec = defRetryMaxElapsedTimeSec
	}

	if conf.Trace.Addr == "" {
		conf.Trace.Addr = defAddr(conf.effectiveProtocol(conf.Trace.Protocol))
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
		conf.Metric.Addr = defAddr(conf.effectiveProtocol(conf.Metric.Protocol))
	}
	if conf.Metric.URLPath == "" {
		conf.Metric.URLPath = defMetricURLPath
	}
	if conf.Metric.AutoRotateTime < 1 {
		conf.Metric.AutoRotateTime = defMetricAutoRotateTime
	}
	if conf.Metric.ExportTimeout < 1 {
		conf.Metric.ExportTimeout = defMetricExportTimeout
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

// effectiveProtocol 返回信号最终生效的传输协议: 信号级配置优先, 其次全局配置, 最后默认值
func (conf *Config) effectiveProtocol(signalProtocol string) string {
	if signalProtocol != "" {
		return signalProtocol
	}
	if conf.Protocol != "" {
		return conf.Protocol
	}
	return defProtocol
}

// mergeHeaders 合并全局与信号级请求头, 键统一转小写(gRPC metadata 要求小写), 信号级覆盖全局
func (conf *Config) mergeHeaders(signal map[string]string) map[string]string {
	if len(conf.Headers) == 0 && len(signal) == 0 {
		return nil
	}
	out := make(map[string]string, len(conf.Headers)+len(signal))
	for k, v := range conf.Headers {
		out[strings.ToLower(strings.TrimSpace(k))] = v
	}
	for k, v := range signal {
		out[strings.ToLower(strings.TrimSpace(k))] = v
	}
	return out
}

// defAddr 返回协议对应的默认地址
func defAddr(protocol string) string {
	if protocol == ProtocolGRPC {
		return defGRPCAddr
	}
	return defHTTPAddr
}

// normalizeProtocol 归一化传输协议, 空值保持不变以便继承上层配置
func normalizeProtocol(protocol string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "":
		return "", nil
	case ProtocolGRPC:
		return ProtocolGRPC, nil
	case ProtocolHTTPProtobuf:
		return ProtocolHTTPProtobuf, nil
	default:
		return "", fmt.Errorf("不支持的传输协议 %q, 可选值: %q, %q", protocol, ProtocolGRPC, ProtocolHTTPProtobuf)
	}
}
