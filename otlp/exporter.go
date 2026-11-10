package otlp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// gzipCompressor gRPC 压缩器名称
const gzipCompressor = "gzip"

// newTraceExporter 按协议创建 trace 导出器
func newTraceExporter(conf TraceConfig, protocol string) (tracesdk.SpanExporter, error) {
	switch protocol {
	case ProtocolGRPC:
		return newGRPCTraceExporter(conf)
	case ProtocolHTTPProtobuf:
		return newHTTPTraceExporter(conf)
	default:
		return nil, fmt.Errorf("不支持的 otlp trace 传输协议 %q", protocol)
	}
}

// newHTTPTraceExporter 创建 OTLP/HTTP trace 导出器
func newHTTPTraceExporter(conf TraceConfig) (tracesdk.SpanExporter, error) {
	compress := otlptracehttp.NoCompression
	if conf.Gzip {
		compress = otlptracehttp.GzipCompression
	}

	return otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpointURL(conf.Addr),
		otlptracehttp.WithURLPath(conf.URLPath),
		otlptracehttp.WithCompression(compress),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         conf.Retry.Enabled,
			InitialInterval: time.Duration(conf.Retry.InitialIntervalSec) * time.Second,
			MaxInterval:     time.Duration(conf.Retry.MaxIntervalSec) * time.Second,
			MaxElapsedTime:  time.Duration(conf.Retry.MaxElapsedTimeSec) * time.Second,
		}),
	)
}

// newGRPCTraceExporter 创建 OTLP/gRPC trace 导出器
func newGRPCTraceExporter(conf TraceConfig) (tracesdk.SpanExporter, error) {
	target, secure, err := parseGRPCEndpoint(conf.Addr)
	if err != nil {
		return nil, err
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(target),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         conf.Retry.Enabled,
			InitialInterval: time.Duration(conf.Retry.InitialIntervalSec) * time.Second,
			MaxInterval:     time.Duration(conf.Retry.MaxIntervalSec) * time.Second,
			MaxElapsedTime:  time.Duration(conf.Retry.MaxElapsedTimeSec) * time.Second,
		}),
	}
	opts = append(opts, grpcSecurityOptions(secure)...)
	if conf.Gzip {
		opts = append(opts, otlptracegrpc.WithCompressor(gzipCompressor))
	}

	return otlptracegrpc.New(context.Background(), opts...)
}

// newMetricExporter 按协议创建 metric 导出器
func newMetricExporter(conf MetricConfig, protocol string) (metricsdk.Exporter, error) {
	switch protocol {
	case ProtocolGRPC:
		return newGRPCMetricExporter(conf)
	case ProtocolHTTPProtobuf:
		return newHTTPMetricExporter(conf)
	default:
		return nil, fmt.Errorf("不支持的 otlp metric 传输协议 %q", protocol)
	}
}

// newHTTPMetricExporter 创建 OTLP/HTTP metric 导出器
func newHTTPMetricExporter(conf MetricConfig) (metricsdk.Exporter, error) {
	compress := otlpmetrichttp.NoCompression
	if conf.Gzip {
		compress = otlpmetrichttp.GzipCompression
	}

	return otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithEndpointURL(conf.Addr),
		otlpmetrichttp.WithURLPath(conf.URLPath),
		otlpmetrichttp.WithCompression(compress),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
			Enabled:         conf.Retry.Enabled,
			InitialInterval: time.Duration(conf.Retry.InitialIntervalSec) * time.Second,
			MaxInterval:     time.Duration(conf.Retry.MaxIntervalSec) * time.Second,
			MaxElapsedTime:  time.Duration(conf.Retry.MaxElapsedTimeSec) * time.Second,
		}),
	)
}

// newGRPCMetricExporter 创建 OTLP/gRPC metric 导出器
func newGRPCMetricExporter(conf MetricConfig) (metricsdk.Exporter, error) {
	target, secure, err := parseGRPCEndpoint(conf.Addr)
	if err != nil {
		return nil, err
	}

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(target),
		otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
			Enabled:         conf.Retry.Enabled,
			InitialInterval: time.Duration(conf.Retry.InitialIntervalSec) * time.Second,
			MaxInterval:     time.Duration(conf.Retry.MaxIntervalSec) * time.Second,
			MaxElapsedTime:  time.Duration(conf.Retry.MaxElapsedTimeSec) * time.Second,
		}),
	}
	opts = append(opts, grpcMetricSecurityOptions(secure)...)
	if conf.Gzip {
		opts = append(opts, otlpmetricgrpc.WithCompressor(gzipCompressor))
	}

	return otlpmetricgrpc.New(context.Background(), opts...)
}

// newLogExporter 按协议创建 log 导出器
func newLogExporter(conf LogConfig, protocol string) (logsdk.Exporter, error) {
	switch protocol {
	case ProtocolGRPC:
		return newGRPCLogExporter(conf)
	case ProtocolHTTPProtobuf:
		return newHTTPLogExporter(conf)
	default:
		return nil, fmt.Errorf("不支持的 otlp log 传输协议 %q", protocol)
	}
}

// newHTTPLogExporter 创建 OTLP/HTTP log 导出器
func newHTTPLogExporter(conf LogConfig) (logsdk.Exporter, error) {
	compress := otlploghttp.NoCompression
	if conf.Gzip {
		compress = otlploghttp.GzipCompression
	}

	return otlploghttp.New(context.Background(),
		otlploghttp.WithEndpointURL(conf.Addr),
		otlploghttp.WithURLPath(conf.URLPath),
		otlploghttp.WithCompression(compress),
		otlploghttp.WithRetry(otlploghttp.RetryConfig{
			Enabled:         conf.Retry.Enabled,
			InitialInterval: time.Duration(conf.Retry.InitialIntervalSec) * time.Second,
			MaxInterval:     time.Duration(conf.Retry.MaxIntervalSec) * time.Second,
			MaxElapsedTime:  time.Duration(conf.Retry.MaxElapsedTimeSec) * time.Second,
		}),
	)
}

// newGRPCLogExporter 创建 OTLP/gRPC log 导出器
func newGRPCLogExporter(conf LogConfig) (logsdk.Exporter, error) {
	target, secure, err := parseGRPCEndpoint(conf.Addr)
	if err != nil {
		return nil, err
	}

	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(target),
		otlploggrpc.WithRetry(otlploggrpc.RetryConfig{
			Enabled:         conf.Retry.Enabled,
			InitialInterval: time.Duration(conf.Retry.InitialIntervalSec) * time.Second,
			MaxInterval:     time.Duration(conf.Retry.MaxIntervalSec) * time.Second,
			MaxElapsedTime:  time.Duration(conf.Retry.MaxElapsedTimeSec) * time.Second,
		}),
	}
	opts = append(opts, grpcLogSecurityOptions(secure)...)
	if conf.Gzip {
		opts = append(opts, otlploggrpc.WithCompressor(gzipCompressor))
	}

	return otlploggrpc.New(context.Background(), opts...)
}

// parseGRPCEndpoint 解析 gRPC 目标地址并补全端口.
// 支持三种写法:
//   - host:port     明文, 端口必填
//   - http://host[:port]  明文, 端口缺省 4317
//   - https://host[:port] TLS, 端口缺省 443
func parseGRPCEndpoint(addr string) (target string, secure bool, err error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", false, fmt.Errorf("otlp grpc 地址不能为空")
	}

	// 无 scheme 时必须显式写端口, 避免退化成 grpc 默认的 443 明文连接
	if !strings.Contains(addr, "://") {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			return "", false, fmt.Errorf("otlp grpc 地址 %q 缺少端口, 应形如 localhost:4317", addr)
		}
		return addr, false, nil
	}

	u, err := url.Parse(addr)
	if err != nil {
		return "", false, fmt.Errorf("解析 otlp grpc 地址 %q 失败: %w", addr, err)
	}
	host := u.Hostname()
	if host == "" {
		return "", false, fmt.Errorf("otlp grpc 地址 %q 缺少 host", addr)
	}
	port := u.Port()

	switch strings.ToLower(u.Scheme) {
	case "http":
		if port == "" {
			port = "4317"
		}
		return net.JoinHostPort(host, port), false, nil
	case "https":
		if port == "" {
			port = "443"
		}
		return net.JoinHostPort(host, port), true, nil
	default:
		return "", false, fmt.Errorf("otlp grpc 地址 %q 不支持协议 %q, 仅支持 host:port、http:// 或 https://", addr, u.Scheme)
	}
}

// grpcSecurityOptions 返回 trace gRPC 导出器的连接安全选项
func grpcSecurityOptions(secure bool) []otlptracegrpc.Option {
	if secure {
		return []otlptracegrpc.Option{otlptracegrpc.WithTLSCredentials(newTLSCredentials())}
	}
	return []otlptracegrpc.Option{otlptracegrpc.WithInsecure()}
}

// grpcMetricSecurityOptions 返回 metric gRPC 导出器的连接安全选项
func grpcMetricSecurityOptions(secure bool) []otlpmetricgrpc.Option {
	if secure {
		return []otlpmetricgrpc.Option{otlpmetricgrpc.WithTLSCredentials(newTLSCredentials())}
	}
	return []otlpmetricgrpc.Option{otlpmetricgrpc.WithInsecure()}
}

// grpcLogSecurityOptions 返回 log gRPC 导出器的连接安全选项
func grpcLogSecurityOptions(secure bool) []otlploggrpc.Option {
	if secure {
		return []otlploggrpc.Option{otlploggrpc.WithTLSCredentials(newTLSCredentials())}
	}
	return []otlploggrpc.Option{otlploggrpc.WithInsecure()}
}

// newTLSCredentials 创建使用系统根证书的 TLS 凭据
func newTLSCredentials() credentials.TransportCredentials {
	return credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
}
